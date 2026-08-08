import { useLocalization } from "../../localization/use-localization.js";
import { useSearchIndex } from "../hooks/useSearchIndex.js";
import SearchResult_default from "./SearchResult.vue.js";
import { Fragment, computed, createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, isRef, openBlock, ref, renderList, toDisplayString, unref, useId, watch, withCtx, withKeys, withModifiers } from "vue";
import { ScalarModal } from "@scalar/components/modal";
import { ScalarSearchInput } from "@scalar/components/search-input";
import { ScalarSearchResultList } from "@scalar/components/search-results";
//#region src/features/Search/components/SearchModal.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	class: "mb-0 flex flex-col",
	role: "search"
};
var _hoisted_2 = {
	"aria-hidden": "true",
	class: "contents"
};
var _hoisted_3 = { class: "sr-only" };
var SearchModal_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SearchModal",
	props: {
		modalState: {},
		document: {},
		eventBus: {},
		modelsSectionLabel: {}
	},
	setup(__props) {
		const props = __props;
		const { translate } = useLocalization();
		/** Base id for the search form */
		const id = useId();
		/** An id for the results listbox */
		const listboxId = `${id}-search-result`;
		/** An id for the results instructions */
		const instructionsId = `${id}-search-instructions`;
		const { query, results } = useSearchIndex(() => props.document, () => props.modelsSectionLabel, () => ({
			heading: translate("search.entryHeading"),
			tagGroup: translate("search.entryTagGroup"),
			webhook: translate("search.entryWebhook"),
			webhooks: translate("navigation.webhooks"),
			introduction: translate("navigation.introduction")
		}));
		const selectedIndex = ref(void 0);
		/** Clear the query value when the modal is opened */
		watch(() => props.modalState.open, (open) => {
			if (open) query.value = "";
		});
		/** Keyboard navigation */
		const navigateSearchResults = (direction) => {
			const offset = direction === "up" ? -1 : 1;
			const length = results.value.length;
			if (typeof selectedIndex.value === "number") selectedIndex.value = (selectedIndex.value + offset + length) % length;
			else selectedIndex.value = offset === -1 ? length - 1 : 0;
		};
		/** Handle the selection of a search result */
		function handleSelect(idx) {
			if (typeof idx !== "number" || !results.value[idx]) return;
			const result = results.value[idx];
			props.modalState.hide();
			props.eventBus.emit("scroll-to:nav-item", { id: result.item.id });
		}
		/**
		* Active descendant id for the search input
		* NOTE: Result items MUST share this id for the aria-activedescendant attribute to work correctly
		*/
		const activeDescendantId = computed(() => {
			const selectedResult = results.value[selectedIndex.value ?? -1];
			return selectedResult ? `search-result-${selectedResult.item.id}` : void 0;
		});
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarModal), {
				"aria-label": unref(translate)("search.label"),
				state: __props.modalState,
				variant: "search"
			}, {
				default: withCtx(() => [
					createElementVNode("div", _hoisted_1, [createVNode(unref(ScalarSearchInput), {
						modelValue: unref(query),
						"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => isRef(query) ? query.value = $event : null),
						"aria-activedescendant": activeDescendantId.value,
						"aria-autocomplete": "list",
						"aria-controls": listboxId,
						"aria-describedby": instructionsId,
						clearLabel: unref(translate)("search.clear"),
						label: unref(translate)("search.inputLabel"),
						placeholder: unref(translate)("search.placeholder"),
						role: "combobox",
						onBlur: _cache[1] || (_cache[1] = ($event) => selectedIndex.value = void 0),
						onKeydown: [
							_cache[2] || (_cache[2] = withKeys(withModifiers(($event) => navigateSearchResults("down"), ["stop", "prevent"]), ["down"])),
							_cache[3] || (_cache[3] = withKeys(withModifiers(() => handleSelect(selectedIndex.value), ["stop", "prevent"]), ["enter"])),
							_cache[4] || (_cache[4] = withKeys(withModifiers(($event) => navigateSearchResults("up"), ["stop", "prevent"]), ["up"]))
						]
					}, null, 8, [
						"modelValue",
						"aria-activedescendant",
						"clearLabel",
						"label",
						"placeholder"
					])]),
					createVNode(unref(ScalarSearchResultList), {
						id: listboxId,
						"aria-label": unref(translate)("search.results"),
						class: "custom-scroll px-1 pb-1",
						noResults: !unref(results).length
					}, {
						query: withCtx(() => [createTextVNode(toDisplayString(unref(query)), 1)]),
						default: withCtx(() => [(openBlock(true), createElementBlock(Fragment, null, renderList(unref(results), (result, idx) => {
							return openBlock(), createBlock(SearchResult_default, {
								id: `search-result-${result.item.id}`,
								key: result.refIndex,
								isSelected: selectedIndex.value === idx,
								modelsSectionLabel: props.modelsSectionLabel,
								result,
								onClick: withModifiers(() => handleSelect(idx), ["prevent"])
							}, null, 8, [
								"id",
								"isSelected",
								"modelsSectionLabel",
								"result",
								"onClick"
							]);
						}), 128))]),
						_: 1
					}, 8, ["aria-label", "noResults"]),
					createElementVNode("div", {
						id: instructionsId,
						class: "ref-search-meta"
					}, [createElementVNode("span", _hoisted_2, [createElementVNode("span", null, "↑↓ " + toDisplayString(unref(translate)("search.navigate")), 1), createElementVNode("span", null, "⏎ " + toDisplayString(unref(translate)("search.select")), 1)]), createElementVNode("span", _hoisted_3, toDisplayString(unref(translate)("search.instructions")), 1)])
				]),
				_: 1
			}, 8, ["aria-label", "state"]);
		};
	}
});
//#endregion
export { SearchModal_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SearchModal.vue.script.js.map