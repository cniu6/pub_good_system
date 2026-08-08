import { Fragment, computed, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref } from "vue";
import { ScalarIconEnvelopeSimple, ScalarIconLink } from "@scalar/icons";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
import { cva } from "@scalar/use-hooks/useBindCx";
//#region src/features/info-object/Contact.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0"
};
var _hoisted_2 = ["href"];
var _hoisted_3 = { class: "ml-1 empty:hidden" };
var _hoisted_4 = {
	key: 1,
	class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0"
};
var _hoisted_5 = ["href"];
var _hoisted_6 = { class: "ml-1 empty:hidden" };
var _hoisted_7 = {
	key: 2,
	class: "group narrow:border-r-0 narrow:first:ml-0 flex items-center border-r first:ml-auto last:border-r-0"
};
var Contact_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Contact",
	props: { value: {} },
	setup(__props) {
		/**
		* The contact URL comes from the OpenAPI document, which is untrusted input, so a protocol like
		* `javascript:` would execute script on click. Fall back to the plain text label in that case.
		*/
		const url = computed(() => sanitizeUrl(__props.value?.url));
		const variants = cva({
			base: "text-c-1 mr-2 flex min-h-7 min-w-7 items-center rounded-lg px-2 py-1 group-last:mr-0 narrow:border",
			variants: { link: { true: "no-underline hover:bg-b-2" } }
		});
		return (_ctx, _cache) => {
			return __props.value ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [__props.value.email ? (openBlock(), createElementBlock("div", _hoisted_1, [createElementVNode("a", {
				class: normalizeClass(unref(variants)({ link: true })),
				href: `mailto:${__props.value.email}`
			}, [createVNode(unref(ScalarIconEnvelopeSimple), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_3, toDisplayString(__props.value.name), 1)], 10, _hoisted_2)])) : createCommentVNode("", true), url.value ? (openBlock(), createElementBlock("div", _hoisted_4, [createElementVNode("a", {
				class: normalizeClass(unref(variants)({ link: true })),
				href: url.value,
				rel: "noopener noreferrer",
				target: "_blank"
			}, [createVNode(unref(ScalarIconLink), {
				class: "size-3 text-current",
				weight: "bold"
			}), createElementVNode("span", _hoisted_6, toDisplayString(__props.value.email ? "" : __props.value.name), 1)], 10, _hoisted_5)])) : !__props.value.email && __props.value.name ? (openBlock(), createElementBlock("div", _hoisted_7, [createElementVNode("span", { class: normalizeClass(unref(variants)({ link: false })) }, toDisplayString(__props.value.name), 3)])) : createCommentVNode("", true)], 64)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Contact_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Contact.vue.script.js.map