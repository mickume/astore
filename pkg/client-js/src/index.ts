/**
 * Zot Artifact Store JavaScript/TypeScript Client SDK
 *
 * A TypeScript client library for interacting with the Zot Artifact Store.
 *
 * @example
 * ```typescript
 * import { Client } from '@astore/client';
 *
 * const client = new Client({
 *   baseURL: 'https://artifacts.example.com',
 *   token: 'your-token'
 * });
 *
 * // Upload artifact
 * await client.upload('mybucket', 'myfile.tar.gz', buffer);
 *
 * // Download artifact
 * const data = await client.download('mybucket', 'myfile.tar.gz');
 * ```
 *
 * @packageDocumentation
 */

export { Client } from './client';
export * from './types';
export * from './exceptions';
