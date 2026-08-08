import { resolveOperationWithTraits } from "@scalar/workspace-store/channel-example";
import { getResolvedRef, mergeSiblingReferences } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/components/Content/AsyncApi/helpers/resolve-async-api-nodes.ts
/**
* Resolve a channel from the document by its `document.channels` key.
*
* Siblings are merged so keys declared alongside a `$ref` are kept rather than dropped, matching
* how the rest of the OpenAPI/AsyncAPI rendering resolves references.
*/
var resolveAsyncApiChannel = (document, channelName) => {
	const node = document.channels?.[channelName];
	return node ? getResolvedRef(node, mergeSiblingReferences) : void 0;
};
/**
* Resolve a message from the channel it lives on. The navigation entry only carries the identifying
* keys, so we walk `document.channels[channelName].messages[messageName]`.
*/
var resolveAsyncApiMessage = (document, channelName, messageName) => {
	const node = resolveAsyncApiChannel(document, channelName)?.messages?.[messageName];
	return node ? getResolvedRef(node, mergeSiblingReferences) : void 0;
};
/**
* Resolve an operation from the document by its `document.operations` key.
*
* Operation traits are merged in (matching the channel connection UI) so trait-only fields render
* as part of the operation.
*/
var resolveAsyncApiOperation = (document, operationName) => {
	const node = document.operations?.[operationName];
	return node ? resolveOperationWithTraits(getResolvedRef(node, mergeSiblingReferences)) : void 0;
};
//#endregion
export { resolveAsyncApiChannel, resolveAsyncApiMessage, resolveAsyncApiOperation };

//# sourceMappingURL=resolve-async-api-nodes.js.map