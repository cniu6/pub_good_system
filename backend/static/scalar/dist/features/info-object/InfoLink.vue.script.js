import { Fragment, computed, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconLink } from "@scalar/icons";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
//#region src/features/info-object/InfoLink.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0" };
var _hoisted_2 = ["href"];
var _hoisted_3 = { class: "ml-1 empty:hidden" };
var _hoisted_4 = { class: "ml-1 empty:hidden" };
var InfoLink_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoLink",
	props: {
		name: {},
		url: {}
	},
	setup(__props) {
		/**
		* The link URL comes from the `x-scalar-links` extension of the OpenAPI document, which is
		* untrusted input, so a protocol like `javascript:` would execute script on click. Fall back to
		* the plain text label in that case.
		*/
		const safeUrl = computed(() => sanitizeUrl(__props.url));
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [safeUrl.value ? (openBlock(), createElementBlock("a", {
				key: 0,
				class: "text-c-1 hover:bg-b-2 narrow:border mr-2 flex min-h-7 min-w-7 items-center rounded-lg px-2 py-1 no-underline group-last:mr-0",
				href: safeUrl.value,
				rel: "noopener noreferrer",
				target: "_blank"
			}, [createVNode(unref(ScalarIconLink), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_3, toDisplayString(__props.name), 1)], 8, _hoisted_2)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createVNode(unref(ScalarIconLink), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_4, toDisplayString(__props.name), 1)], 64))]);
		};
	}
});
//#endregion
export { InfoLink_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoLink.vue.script.js.map