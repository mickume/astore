/**
 * Exception classes for Zot Artifact Store client
 */

/**
 * Base exception for all Artifact Store errors
 */
export class ArtifactStoreError extends Error {
  public statusCode?: number;

  constructor(message: string, statusCode?: number) {
    super(message);
    this.name = 'ArtifactStoreError';
    this.statusCode = statusCode;
    Object.setPrototypeOf(this, ArtifactStoreError.prototype);
  }
}

/**
 * 400 Bad Request
 */
export class BadRequestError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 400);
    this.name = 'BadRequestError';
    Object.setPrototypeOf(this, BadRequestError.prototype);
  }
}

/**
 * 401 Unauthorized
 */
export class UnauthorizedError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 401);
    this.name = 'UnauthorizedError';
    Object.setPrototypeOf(this, UnauthorizedError.prototype);
  }
}

/**
 * 403 Forbidden
 */
export class ForbiddenError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 403);
    this.name = 'ForbiddenError';
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}

/**
 * 404 Not Found
 */
export class NotFoundError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 404);
    this.name = 'NotFoundError';
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * 409 Conflict
 */
export class ConflictError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 409);
    this.name = 'ConflictError';
    Object.setPrototypeOf(this, ConflictError.prototype);
  }
}

/**
 * 500 Internal Server Error
 */
export class InternalServerError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 500);
    this.name = 'InternalServerError';
    Object.setPrototypeOf(this, InternalServerError.prototype);
  }
}

/**
 * 503 Service Unavailable
 */
export class ServiceUnavailableError extends ArtifactStoreError {
  constructor(message: string) {
    super(message, 503);
    this.name = 'ServiceUnavailableError';
    Object.setPrototypeOf(this, ServiceUnavailableError.prototype);
  }
}

/**
 * Raise appropriate exception for HTTP status code
 */
export function raiseForStatus(statusCode: number, message: string): void {
  const errorMap: Record<number, new (msg: string) => ArtifactStoreError> = {
    400: BadRequestError,
    401: UnauthorizedError,
    403: ForbiddenError,
    404: NotFoundError,
    409: ConflictError,
    500: InternalServerError,
    503: ServiceUnavailableError,
  };

  const ErrorClass = errorMap[statusCode] || ArtifactStoreError;
  throw new ErrorClass(message);
}
