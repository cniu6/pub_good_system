import { useLocalization } from "../localization/use-localization.js";
import { computed, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconScroll } from "@scalar/icons";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
//#region src/features/info-object/TermsOfService.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0"
};
var _hoisted_2 = ["href"];
var _hoisted_3 = { class: "ml-1 empty:hidden" };
var TermsOfService_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "TermsOfService",
	props: { value: {} },
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* The terms of service URL comes from the OpenAPI document, which is untrusted input, so a
		* protocol like `javascript:` would execute script on click. Hide the link in that case.
		*/
		const url = computed(() => sanitizeUrl(__props.value));
		return (_ctx, _cache) => {
			return url.value ? (openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("a", {
				class: "text-c-1 hover:bg-b-2 narrow:border mr-2 flex min-h-7 min-w-7 items-center rounded-lg px-2 py-1 no-underline group-last:mr-0",
				href: url.value,
				rel: "noopener noreferrer",
				target: "_blank"
			}, [createVNode(unref(ScalarIconScroll), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_3, toDisplayString(unref(translate)("info.termsOfService")), 1)], 8, _hoisted_2)])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { TermsOfService_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=TermsOfService.vue.script.js.map