import { createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, ref, renderSlot, unref, withCtx } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { useElementHover } from "@vueuse/core";
import { ScalarIconCaretRight } from "@scalar/icons";
//#region src/components/Section/SectionAccordion.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "section-accordion-button-content" };
var _hoisted_2 = {
	key: 0,
	class: "section-accordion-button-actions"
};
var _hoisted_3 = {
	key: 0,
	class: "section-accordion-description"
};
var _hoisted_4 = { class: "section-accordion-content-card" };
var SectionAccordion_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SectionAccordion",
	props: {
		transparent: { type: Boolean },
		modelValue: { type: Boolean }
	},
	emits: ["update:modelValue"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		const button = ref();
		const isHovered = useElementHover(button);
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(Disclosure), {
				as: "section",
				class: normalizeClass(["section-accordion", { "section-accordion-transparent": __props.transparent }])
			}, {
				default: withCtx(() => [createVNode(unref(DisclosureButton), {
					ref_key: "button",
					ref: button,
					class: "section-accordion-button",
					onClick: _cache[0] || (_cache[0] = () => emit("update:modelValue", !__props.modelValue))
				}, {
					default: withCtx(() => [
						createElementVNode("div", _hoisted_1, [renderSlot(_ctx.$slots, "title", {}, void 0, true)]),
						_ctx.$slots.actions ? (openBlock(), createElementBlock("div", _hoisted_2, [renderSlot(_ctx.$slots, "actions", { active: unref(isHovered) || __props.modelValue }, void 0, true)])) : createCommentVNode("", true),
						createVNode(unref(ScalarIconCaretRight), { class: normalizeClass(["section-accordion-chevron size-4.5 transition-transform", { "rotate-90": __props.modelValue }]) }, null, 8, ["class"])
					]),
					_: 3
				}, 512), __props.modelValue ? (openBlock(), createBlock(unref(DisclosurePanel), {
					key: 0,
					class: "section-accordion-content",
					static: ""
				}, {
					default: withCtx(() => [_ctx.$slots.description ? (openBlock(), createElementBlock("div", _hoisted_3, [renderSlot(_ctx.$slots, "description", {}, void 0, true)])) : createCommentVNode("", true), createElementVNode("div", _hoisted_4, [renderSlot(_ctx.$slots, "default", {}, void 0, true)])]),
					_: 3
				})) : createCommentVNode("", true)]),
				_: 3
			}, 8, ["class"]);
		};
	}
});
//#endregion
export { SectionAccordion_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SectionAccordion.vue.script.js.map