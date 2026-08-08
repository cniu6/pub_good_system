import type { ClientPlugin } from '@scalar/oas-utils/helpers';
import type { ApiReferenceConfiguration } from '@scalar/types/api-reference';
import type { XScalarEnvironment } from '@scalar/workspace-store/schemas/extensions/document/x-scalar-environments';
import { type ComputedRef } from 'vue';
/**
 * Maps API reference configuration callbacks to client plugins.
 *
 * This function transforms the onBeforeRequest, onRequestBuilt, and onRequestSent
 * callbacks into the new plugin hook system. The mapping is reactive, so changes
 * to the configuration will automatically update the plugin hooks.
 *
 * Note: onRequestBuilt is mapped to the requestBuilt hook so the callback receives
 * the exact fetch Request that is sent over the wire. Mutating its headers modifies
 * the outgoing request, and hashing its body produces a hash that matches what the
 * server receives (a rebuilt multipart body would get a different boundary).
 *
 * Note: onRequestSent is mapped to responseReceived hook. This is not a perfect
 * one-to-one mapping, but it maintains backward compatibility with the old API.
 * The old callback receives only the URL string, while the new hook receives
 * the full response object.
 *
 * @param config - Reactive configuration object containing optional hook callbacks
 * @returns Array containing a single plugin with the mapped hooks
 */
export declare const mapConfigPlugins: (config: ComputedRef<ApiReferenceConfiguration>, environment: ComputedRef<XScalarEnvironment>) => ClientPlugin[];
//# sourceMappingURL=map-config-plugins.d.ts.map