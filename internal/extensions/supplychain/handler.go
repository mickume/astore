package supplychain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	scPkg "github.com/candlekeep/zot-artifact-store/internal/supplychain"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/log"
)

// Handler handles supply chain security API requests
type Handler struct {
	metadataStore *storage.MetadataStore
	signer        *scPkg.Signer
	logger        log.Logger
}

// NewHandler creates a new supply chain handler
func NewHandler(metadataStore *storage.MetadataStore, signer *scPkg.Signer, logger log.Logger) *Handler {
	return &Handler{
		metadataStore: metadataStore,
		signer:        signer,
		logger:        logger,
	}
}

// RegisterRoutes registers supply chain security API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Signature operations
	router.HandleFunc("/supplychain/sign/{bucket}/{key:.*}", h.SignArtifact).Methods("POST")
	router.HandleFunc("/supplychain/signatures/{bucket}/{key:.*}", h.GetSignatures).Methods("GET")
	router.HandleFunc("/supplychain/verify/{bucket}/{key:.*}", h.VerifyArtifact).Methods("POST")

	// SBOM operations
	router.HandleFunc("/supplychain/sbom/{bucket}/{key:.*}", h.AttachSBOM).Methods("POST")
	router.HandleFunc("/supplychain/sbom/{bucket}/{key:.*}", h.GetSBOM).Methods("GET")

	// Attestation operations
	router.HandleFunc("/supplychain/attestations/{bucket}/{key:.*}", h.AddAttestation).Methods("POST")
	router.HandleFunc("/supplychain/attestations/{bucket}/{key:.*}", h.GetAttestations).Methods("GET")
}

// === Signature Operations ===

// SignArtifactRequest represents a request to sign an artifact
type SignArtifactRequest struct {
	SignedBy string `json:"signedBy"`
}

// SignArtifact signs an artifact and stores the signature
func (h *Handler) SignArtifact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	var req SignArtifactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SignedBy == "" {
		req.SignedBy = "system"
	}

	// Get artifact metadata
	artifact, err := h.metadataStore.GetArtifact(bucket, key)
	if err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}

	// Read artifact data for signing
	// In a real implementation, this would read from storage
	// For now, we'll use the artifact's digest as the data
	data := []byte(artifact.Digest.String())

	// Sign the artifact
	signature, err := h.signer.SignArtifact(artifactID, data, req.SignedBy)
	if err != nil {
		h.logger.Error().Err(err).Str("artifactId", artifactID).Msg("failed to sign artifact")
		http.Error(w, "Failed to sign artifact", http.StatusInternalServerError)
		return
	}

	// Store signature
	if err := h.metadataStore.StoreSignature(signature); err != nil {
		h.logger.Error().Err(err).Msg("failed to store signature")
		http.Error(w, "Failed to store signature", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("artifactId", artifactID).Str("signatureId", signature.ID).Msg("artifact signed")

	h.writeJSON(w, http.StatusCreated, signature)
}

// GetSignatures retrieves all signatures for an artifact
func (h *Handler) GetSignatures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	signatures, err := h.metadataStore.ListSignaturesForArtifact(artifactID)
	if err != nil {
		h.logger.Error().Err(err).Str("artifactId", artifactID).Msg("failed to list signatures")
		http.Error(w, "Failed to list signatures", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"artifactId": artifactID,
		"signatures": signatures,
		"count":      len(signatures),
	})
}

// VerifyArtifact verifies all signatures for an artifact
func (h *Handler) VerifyArtifact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	// Get artifact
	artifact, err := h.metadataStore.GetArtifact(bucket, key)
	if err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}

	// Get signatures
	signatures, err := h.metadataStore.ListSignaturesForArtifact(artifactID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list signatures")
		http.Error(w, "Failed to list signatures", http.StatusInternalServerError)
		return
	}

	if len(signatures) == 0 {
		http.Error(w, "No signatures found for artifact", http.StatusNotFound)
		return
	}

	// Verify each signature
	data := []byte(artifact.Digest.String())
	results := make([]*models.VerificationResult, 0, len(signatures))

	for _, sig := range signatures {
		result, err := scPkg.VerifySignature(sig, data)
		if err != nil {
			h.logger.Error().Err(err).Str("signatureId", sig.ID).Msg("failed to verify signature")
			continue
		}
		results = append(results, result)
	}

	// Determine overall verification status
	allVerified := len(results) > 0
	for _, result := range results {
		if !result.Verified {
			allVerified = false
			break
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"artifactId":  artifactID,
		"verified":    allVerified,
		"results":     results,
		"totalSigs":   len(signatures),
		"verifiedSigs": countVerified(results),
	})
}

// === SBOM Operations ===

// AttachSBOMRequest represents a request to attach an SBOM
type AttachSBOMRequest struct {
	Format      models.SBOMFormat `json:"format"`
	Version     string            `json:"version"`
	Content     string            `json:"content"` // JSON or XML string
	ContentType string            `json:"contentType"`
	CreatedBy   string            `json:"createdBy"`
}

// AttachSBOM attaches an SBOM to an artifact
func (h *Handler) AttachSBOM(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	var req AttachSBOMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate artifact exists
	_, err := h.metadataStore.GetArtifact(bucket, key)
	if err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}

	// Calculate hash of content
	hash := sha256.Sum256([]byte(req.Content))
	hashStr := hex.EncodeToString(hash[:])

	// Create SBOM
	sbom := &models.SBOM{
		ID:          uuid.New().String(),
		ArtifactID:  artifactID,
		Format:      req.Format,
		Version:     req.Version,
		Content:     []byte(req.Content),
		ContentType: req.ContentType,
		Hash:        hashStr,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
	}

	// Store SBOM
	if err := h.metadataStore.StoreSBOM(sbom); err != nil {
		h.logger.Error().Err(err).Msg("failed to store SBOM")
		http.Error(w, "Failed to store SBOM", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("artifactId", artifactID).Str("sbomId", sbom.ID).Str("format", string(req.Format)).Msg("SBOM attached")

	// Return SBOM without content in response
	response := map[string]interface{}{
		"id":          sbom.ID,
		"artifactId":  sbom.ArtifactID,
		"format":      sbom.Format,
		"version":     sbom.Version,
		"contentType": sbom.ContentType,
		"hash":        sbom.Hash,
		"createdBy":   sbom.CreatedBy,
		"createdAt":   sbom.CreatedAt,
		"size":        len(sbom.Content),
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetSBOM retrieves the SBOM for an artifact
func (h *Handler) GetSBOM(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	sbom, err := h.metadataStore.GetSBOMForArtifact(artifactID)
	if err != nil {
		http.Error(w, "SBOM not found", http.StatusNotFound)
		return
	}

	// Return full SBOM with content
	h.writeJSON(w, http.StatusOK, sbom)
}

// === Attestation Operations ===

// AddAttestationRequest represents a request to add an attestation
type AddAttestationRequest struct {
	Type          models.AttestationType `json:"type"`
	Predicate     map[string]interface{} `json:"predicate"`
	PredicateType string                 `json:"predicateType"`
	CreatedBy     string                 `json:"createdBy"`
}

// AddAttestation adds an attestation to an artifact
func (h *Handler) AddAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	var req AddAttestationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate artifact exists
	_, err := h.metadataStore.GetArtifact(bucket, key)
	if err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}

	// Create attestation
	attestation := &models.Attestation{
		ID:            uuid.New().String(),
		ArtifactID:    artifactID,
		Type:          req.Type,
		Predicate:     req.Predicate,
		PredicateType: req.PredicateType,
		CreatedBy:     req.CreatedBy,
		CreatedAt:     time.Now(),
	}

	// Store attestation
	if err := h.metadataStore.StoreAttestation(attestation); err != nil {
		h.logger.Error().Err(err).Msg("failed to store attestation")
		http.Error(w, "Failed to store attestation", http.StatusInternalServerError)
		return
	}

	h.logger.Info().
		Str("artifactId", artifactID).
		Str("attestationId", attestation.ID).
		Str("type", string(req.Type)).
		Msg("attestation added")

	h.writeJSON(w, http.StatusCreated, attestation)
}

// GetAttestations retrieves all attestations for an artifact
func (h *Handler) GetAttestations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	artifactID := bucket + "/" + key

	attestations, err := h.metadataStore.ListAttestationsForArtifact(artifactID)
	if err != nil {
		h.logger.Error().Err(err).Str("artifactId", artifactID).Msg("failed to list attestations")
		http.Error(w, "Failed to list attestations", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"artifactId":   artifactID,
		"attestations": attestations,
		"count":        len(attestations),
	})
}

// === Helper Functions ===

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func countVerified(results []*models.VerificationResult) int {
	count := 0
	for _, r := range results {
		if r.Verified {
			count++
		}
	}
	return count
}
