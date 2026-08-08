import { ALL, createReachabilityContext, getOperationReachability } from "@scalar/workspace-store/channel-example";
import { getResolvedRef, mergeSiblingReferences } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/blocks/scalar-asyncapi-sidebar-filters-block/helpers/filter-async-api-navigation.ts
/** Whether the filter would keep every entry, so the tree can be returned untouched. */
var isNoopFilter = ({ protocol, server }) => (!protocol || protocol === ALL) && (!server || server === ALL);
/** Whether a selection (protocol or server) keeps an operation, given the set it is reachable through. */
var selectionMatches = (reachable, selected) => !selected || selected === ALL || reachable.has(selected);
/**
* Filters one navigation entry against the selected protocol/server.
*
* - `asyncapi-operation` — kept only when the operation matches both filters.
* - `asyncapi-channel` / `tag` — recursed into, then dropped when they have no
*   children left (so empty channels and tags disappear from the sidebar).
* - Everything else (description, models, schemas) passes through unchanged.
*
* `context` carries the document-level lookups so they are built once per filter
* pass rather than recomputed for every operation.
*
* Returns `null` when the entry should be removed.
*/
var filterEntry = (entry, document, filter, context) => {
	if (entry.type === "asyncapi-operation") {
		const operationNode = document.operations?.[entry.operationName];
		if (!operationNode) return entry;
		const { protocols, serverNames } = getOperationReachability(document, getResolvedRef(operationNode, mergeSiblingReferences), context);
		return selectionMatches(protocols, filter.protocol) && selectionMatches(serverNames, filter.server) ? entry : null;
	}
	if (entry.type === "asyncapi-channel" || entry.type === "tag") {
		const originalChildren = entry.children ?? [];
		const children = originalChildren.flatMap((child) => {
			const filtered = filterEntry(child, document, filter, context);
			return filtered ? [filtered] : [];
		});
		if (originalChildren.length > 0 && children.length === 0) return null;
		return {
			...entry,
			children
		};
	}
	return entry;
};
/**
* Filters the AsyncAPI sidebar tree by the selected protocol and/or server.
*
* Returns the original entries untouched when no filter is active, so OpenAPI
* documents and the unfiltered AsyncAPI case pay no cost.
*/
var filterAsyncApiNavigation = (entries, document, filter) => {
	if (isNoopFilter(filter)) return entries;
	const context = createReachabilityContext(document);
	return entries.flatMap((entry) => {
		const filtered = filterEntry(entry, document, filter, context);
		return filtered ? [filtered] : [];
	});
};
//#endregion
export { filterAsyncApiNavigation };

//# sourceMappingURL=filter-async-api-navigation.js.map