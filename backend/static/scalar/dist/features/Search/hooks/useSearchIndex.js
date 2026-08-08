import { createFuseInstance } from "../helpers/create-fuse-instance.js";
import { createSearchIndex } from "../helpers/create-search-index.js";
import { computed, ref, toValue } from "vue";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
//#region src/features/Search/hooks/useSearchIndex.ts
var MAX_SEARCH_RESULTS = 25;
/**
* Creates the search index from an OpenAPI or AsyncAPI document.
*/
function useSearchIndex(document, modelsSectionLabel = DEFAULT_MODELS_SECTION_LABEL, labels = {
	heading: "Heading",
	tagGroup: "Tag Group",
	webhook: "Webhook",
	webhooks: "Webhooks",
	introduction: "Introduction"
}) {
	const searchIndex = computed(() => createSearchIndex(toValue(document), {
		labels: toValue(labels),
		modelsSectionLabel: toValue(modelsSectionLabel) ?? DEFAULT_MODELS_SECTION_LABEL
	}));
	/** When the document changes we replace the search index */
	const fuse = computed(() => {
		const instance = createFuseInstance();
		instance.setCollection(searchIndex.value);
		return instance;
	});
	const query = ref("");
	return {
		results: computed(() => {
			if (query.value.length !== 0) return fuse.value.search(query.value, { limit: MAX_SEARCH_RESULTS });
			return searchIndex.value.slice(0, MAX_SEARCH_RESULTS).map((item, index) => ({
				item,
				refIndex: index
			}));
		}),
		query
	};
}
//#endregion
export { useSearchIndex };

//# sourceMappingURL=useSearchIndex.js.map