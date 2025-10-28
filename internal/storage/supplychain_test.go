package storage_test

import (
	"os"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestSupplyChainStorage(t *testing.T) {
	t.Run("Store and retrieve signature", func(t *testing.T) {
		// Given: A metadata store and a signature
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, err := storage.NewMetadataStore(tmpFile.Name())
		test.AssertNoError(t, err, "creating store")
		defer store.Close()

		signature := &models.Signature{
			ID:         "sig-1",
			ArtifactID: "bucket/artifact",
			Algorithm:  "RSA-SHA256",
			Signature:  []byte("signature-data"),
			PublicKey:  "public-key-pem",
			SignedBy:   "user@example.com",
			SignedAt:   time.Now(),
		}

		// When: Storing the signature
		err = store.StoreSignature(signature)
		test.AssertNoError(t, err, "storing signature")

		// Then: Signature can be retrieved
		retrieved, err := store.GetSignature("sig-1")
		test.AssertNoError(t, err, "retrieving signature")
		test.AssertEqual(t, signature.ID, retrieved.ID, "signature ID")
		test.AssertEqual(t, signature.ArtifactID, retrieved.ArtifactID, "artifact ID")
		test.AssertEqual(t, signature.SignedBy, retrieved.SignedBy, "signed by")
	})

	t.Run("List signatures for artifact", func(t *testing.T) {
		// Given: A metadata store with multiple signatures
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		artifactID := "bucket/artifact"

		sig1 := &models.Signature{ID: "sig-1", ArtifactID: artifactID, SignedBy: "user1"}
		sig2 := &models.Signature{ID: "sig-2", ArtifactID: artifactID, SignedBy: "user2"}
		sig3 := &models.Signature{ID: "sig-3", ArtifactID: "other/artifact", SignedBy: "user3"}

		store.StoreSignature(sig1)
		store.StoreSignature(sig2)
		store.StoreSignature(sig3)

		// When: Listing signatures for artifact
		signatures, err := store.ListSignaturesForArtifact(artifactID)

		// Then: Only relevant signatures are returned
		test.AssertNoError(t, err, "listing signatures")
		test.AssertEqual(t, 2, len(signatures), "signature count")
	})

	t.Run("Store and retrieve SBOM", func(t *testing.T) {
		// Given: A metadata store and an SBOM
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		sbom := &models.SBOM{
			ID:          "sbom-1",
			ArtifactID:  "bucket/artifact",
			Format:      models.SBOMFormatSPDX,
			Version:     "2.3",
			Content:     []byte(`{"spdxVersion":"SPDX-2.3"}`),
			ContentType: "application/json",
			Hash:        "abc123",
			CreatedBy:   "scanner",
			CreatedAt:   time.Now(),
		}

		// When: Storing the SBOM
		err := store.StoreSBOM(sbom)
		test.AssertNoError(t, err, "storing SBOM")

		// Then: SBOM can be retrieved
		retrieved, err := store.GetSBOM("sbom-1")
		test.AssertNoError(t, err, "retrieving SBOM")
		test.AssertEqual(t, sbom.ID, retrieved.ID, "SBOM ID")
		test.AssertEqual(t, sbom.Format, retrieved.Format, "SBOM format")
		test.AssertEqual(t, sbom.Version, retrieved.Version, "SBOM version")
	})

	t.Run("Get SBOM for artifact", func(t *testing.T) {
		// Given: A metadata store with an SBOM
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		sbom := &models.SBOM{
			ID:         "sbom-1",
			ArtifactID: "bucket/artifact",
			Format:     models.SBOMFormatCycloneDX,
			Content:    []byte(`{"bomFormat":"CycloneDX"}`),
		}
		store.StoreSBOM(sbom)

		// When: Getting SBOM for artifact
		retrieved, err := store.GetSBOMForArtifact("bucket/artifact")

		// Then: SBOM is retrieved
		test.AssertNoError(t, err, "retrieving SBOM")
		test.AssertEqual(t, sbom.ID, retrieved.ID, "SBOM ID")
		test.AssertEqual(t, models.SBOMFormatCycloneDX, retrieved.Format, "format")
	})

	t.Run("Store and retrieve attestation", func(t *testing.T) {
		// Given: A metadata store and an attestation
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		attestation := &models.Attestation{
			ID:         "att-1",
			ArtifactID: "bucket/artifact",
			Type:       models.AttestationTypeBuild,
			Predicate: map[string]interface{}{
				"builder": "github-actions",
				"commit":  "abc123",
			},
			PredicateType: "https://slsa.dev/provenance/v0.2",
			CreatedBy:     "ci-system",
			CreatedAt:     time.Now(),
		}

		// When: Storing the attestation
		err := store.StoreAttestation(attestation)
		test.AssertNoError(t, err, "storing attestation")

		// Then: Attestation can be retrieved
		retrieved, err := store.GetAttestation("att-1")
		test.AssertNoError(t, err, "retrieving attestation")
		test.AssertEqual(t, attestation.ID, retrieved.ID, "attestation ID")
		test.AssertEqual(t, attestation.Type, retrieved.Type, "attestation type")
		test.AssertEqual(t, attestation.PredicateType, retrieved.PredicateType, "predicate type")
	})

	t.Run("List attestations for artifact", func(t *testing.T) {
		// Given: A metadata store with multiple attestations
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		artifactID := "bucket/artifact"

		att1 := &models.Attestation{ID: "att-1", ArtifactID: artifactID, Type: models.AttestationTypeBuild}
		att2 := &models.Attestation{ID: "att-2", ArtifactID: artifactID, Type: models.AttestationTypeTest}
		att3 := &models.Attestation{ID: "att-3", ArtifactID: "other/artifact", Type: models.AttestationTypeScan}

		store.StoreAttestation(att1)
		store.StoreAttestation(att2)
		store.StoreAttestation(att3)

		// When: Listing attestations for artifact
		attestations, err := store.ListAttestationsForArtifact(artifactID)

		// Then: Only relevant attestations are returned
		test.AssertNoError(t, err, "listing attestations")
		test.AssertEqual(t, 2, len(attestations), "attestation count")
	})

	t.Run("Delete signature", func(t *testing.T) {
		// Given: A metadata store with a signature
		tmpFile, _ := os.CreateTemp("", "supplychain-test-*.db")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		store, _ := storage.NewMetadataStore(tmpFile.Name())
		defer store.Close()

		signature := &models.Signature{ID: "sig-1", ArtifactID: "bucket/artifact"}
		store.StoreSignature(signature)

		// When: Deleting the signature
		err := store.DeleteSignature("sig-1")
		test.AssertNoError(t, err, "deleting signature")

		// Then: Signature is deleted
		_, err = store.GetSignature("sig-1")
		test.AssertError(t, err, "signature should not exist")
	})
}
