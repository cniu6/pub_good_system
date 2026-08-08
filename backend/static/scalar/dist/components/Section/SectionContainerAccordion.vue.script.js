import { createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, renderSlot, unref, withCtx } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ScalarIconCaretRight } from "@scalar/icons";
//#region src/components/Section/SectionContainerAccordion.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "section-accordion-wrapper" };
var _hoisted_2 = { class: "section-accordion-title" };
var SectionContainerAccordion_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SectionContainerAccordion",
	props: { modelValue: { type: Boolean } },
	emits: ["update:modelValue"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("section", _hoisted_1, [createVNode(unref(Disclosure), {
				as: "div",
				class: "section-accordion"
			}, {
				default: withCtx(() => [createVNode(unref(DisclosureButton), {
					class: "section-accordion-button",
					onClick: _cache[0] || (_cache[0] = ($event) => emit("update:modelValue", !__props.modelValue))
				}, {
					default: withCtx(() => [createVNode(unref(ScalarIconCaretRight), { class: normalizeClass(["section-accordion-chevron size-5 transition-transform", { "rotate-90": __props.modelValue }]) }, null, 8, ["class"]), createElementVNode("div", _hoisted_2, [renderSlot(_ctx.$slots, "title", {}, void 0, true)])]),
					_: 3
				}), __props.modelValue ? (openBlock(), createBlock(unref(DisclosurePanel), {
					key: 0,
					class: "section-accordion-content",
					static: ""
				}, {
					default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, void 0, true)]),
					_: 3
				})) : createCommentVNode("", true)]),
				_: 3
			})]);
		};
	}
});
//#endregion
export { SectionContainerAccordion_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SectionContainerAccordion.vue.script.js.map