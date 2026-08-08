import type { AsyncApiChannelObject, AsyncApiDocument, AsyncApiMessageObject, AsyncApiOperationObject } from '@scalar/types/asyncapi/3.1';
/**
 * Resolve a channel from the document by its `document.channels` key.
 *
 * Siblings are merged so keys declared alongside a `$ref` are kept rather than dropped, matching
 * how the rest of the OpenAPI/AsyncAPI rendering resolves references.
 */
export declare const resolveAsyncApiChannel: (document: AsyncApiDocument, channelName: string) => AsyncApiChannelObject | undefined;
/**
 * Resolve a message from the channel it lives on. The navigation entry only carries the identifying
 * keys, so we walk `document.channels[channelName].messages[messageName]`.
 */
export declare const resolveAsyncApiMessage: (document: AsyncApiDocument, channelName: string, messageName: string) => AsyncApiMessageObject | undefined;
/**
 * Resolve an operation from the document by its `document.operations` key.
 *
 * Operation traits are merged in (matching the channel connection UI) so trait-only fields render
 * as part of the operation.
 */
export declare const resolveAsyncApiOperation: (document: AsyncApiDocument, operationName: string) => AsyncApiOperationObject | undefined;
//# sourceMappingURL=resolve-async-api-nodes.d.ts.map