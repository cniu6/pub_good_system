import { useLocalization } from "../../localization/use-localization.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createSlots, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref, withCtx } from "vue";
import { HttpMethod } from "@scalar/sidebar";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
import { ScalarIconBracketsCurly, ScalarIconTag, ScalarIconTerminalWindow, ScalarIconTextAlignLeft } from "@scalar/icons";
import { ScalarSearchResultItem } from "@scalar/components/search-results";
//#region src/features/Search/components/SearchResult.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "sr-only" };
var _hoisted_2 = { class: "inline-flex items-center gap-1" };
var _hoisted_3 = { class: "sr-only" };
var _hoisted_4 = { class: "sr-only" };
var _hoisted_5 = { class: "sr-only" };
var SearchResult_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SearchResult",
	props: {
		id: {},
		isSelected: { type: Boolean },
		result: {},
		modelsSectionLabel: { default: () => DEFAULT_MODELS_SECTION_LABEL }
	},
	setup(__props) {
		const { translate } = useLocalization();
		const ENTRY_ICONS = {
			heading: ScalarIconTextAlignLeft,
			model: ScalarIconBracketsCurly,
			operation: ScalarIconTerminalWindow,
			tag: ScalarIconTag,
			webhook: ScalarIconTerminalWindow
		};
		const entryLabels = computed(() => ({
			heading: translate("search.entryHeading"),
			operation: translate("search.entryOperation"),
			tag: translate("search.entryTag"),
			model: __props.modelsSectionLabel,
			webhook: translate("search.entryWebhook")
		}));
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarSearchResultItem), {
				id: __props.id,
				icon: ENTRY_ICONS[__props.result.item.type],
				selected: __props.isSelected
			}, createSlots({
				default: withCtx(() => [createElementVNode("span", { class: normalizeClass({ "text-decoration-line": __props.result.item.entry.type === "operation" && __props.result.item.entry.isDeprecated }) }, [
					createElementVNode("span", _hoisted_1, [createTextVNode(toDisplayString(entryLabels.value[__props.result.item.type]) + ":\xA0 ", 1), __props.result.item.entry.type === "operation" && __props.result.item.entry.isDeprecated ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createTextVNode(" (" + toDisplayString(unref(translate)("common.deprecated")) + ")\xA0 ", 1)], 64)) : createCommentVNode("", true)]),
					createTextVNode(" " + toDisplayString(__props.result.item.title) + " ", 1),
					_cache[0] || (_cache[0] = createElementVNode("span", { class: "sr-only" }, ",", -1))
				], 2)]),
				_: 2
			}, [__props.result.item.type !== "webhook" && (__props.result.item.method || __props.result.item.path) && __props.result.item.path !== __props.result.item.title ? {
				name: "description",
				fn: withCtx(() => [createElementVNode("span", _hoisted_2, [
					__props.result.item.type === "operation" ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createVNode(unref(HttpMethod), {
						"aria-hidden": "true",
						method: __props.result.item.method ?? "get"
					}, null, 8, ["method"]), createElementVNode("span", _hoisted_3, toDisplayString(unref(translate)("common.httpMethod")) + ": " + toDisplayString(__props.result.item.method ?? "get"), 1)], 64)) : createCommentVNode("", true),
					createElementVNode("span", _hoisted_4, toDisplayString(unref(translate)("common.path")) + ":\xA0", 1),
					createTextVNode(" " + toDisplayString(__props.result.item.path), 1)
				])]),
				key: "0"
			} : __props.result.item.description ? {
				name: "description",
				fn: withCtx(() => [createElementVNode("span", _hoisted_5, toDisplayString(unref(translate)("common.description")) + ":\xA0", 1), createTextVNode(" " + toDisplayString(__props.result.item.description), 1)]),
				key: "1"
			} : void 0]), 1032, [
				"id",
				"icon",
				"selected"
			]);
		};
	}
});
//#endregion
export { SearchResult_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SearchResult.vue.script.js.map