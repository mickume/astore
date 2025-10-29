"""
Exception classes for Zot Artifact Store client
"""


class ArtifactStoreError(Exception):
    """Base exception for all Artifact Store errors"""

    def __init__(self, message: str, status_code: int = None):
        super().__init__(message)
        self.message = message
        self.status_code = status_code


class BadRequestError(ArtifactStoreError):
    """400 Bad Request"""

    def __init__(self, message: str):
        super().__init__(message, 400)


class UnauthorizedError(ArtifactStoreError):
    """401 Unauthorized"""

    def __init__(self, message: str):
        super().__init__(message, 401)


class ForbiddenError(ArtifactStoreError):
    """403 Forbidden"""

    def __init__(self, message: str):
        super().__init__(message, 403)


class NotFoundError(ArtifactStoreError):
    """404 Not Found"""

    def __init__(self, message: str):
        super().__init__(message, 404)


class ConflictError(ArtifactStoreError):
    """409 Conflict"""

    def __init__(self, message: str):
        super().__init__(message, 409)


class InternalServerError(ArtifactStoreError):
    """500 Internal Server Error"""

    def __init__(self, message: str):
        super().__init__(message, 500)


class ServiceUnavailableError(ArtifactStoreError):
    """503 Service Unavailable"""

    def __init__(self, message: str):
        super().__init__(message, 503)


def raise_for_status(response):
    """Raise appropriate exception for HTTP status code"""
    if response.status_code < 400:
        return

    try:
        error_data = response.json()
        message = error_data.get("error", response.text)
    except Exception:
        message = response.text or f"HTTP {response.status_code}"

    error_map = {
        400: BadRequestError,
        401: UnauthorizedError,
        403: ForbiddenError,
        404: NotFoundError,
        409: ConflictError,
        500: InternalServerError,
        503: ServiceUnavailableError,
    }

    error_class = error_map.get(response.status_code, ArtifactStoreError)
    raise error_class(message)
