import Section_default from "./Section.vue.js";
import Anchor_default from "../Anchor/Anchor.vue.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, renderSlot, unref, useId, withCtx } from "vue";
import { ScalarIconCaretRight } from "@scalar/icons";
//#region src/components/Section/CompactSection.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var _hoisted_2 = [
	"id",
	"aria-controls",
	"aria-expanded",
	"aria-labelledby"
];
var _hoisted_3 = ["id"];
var CompactSection_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "CompactSection",
	props: {
		id: {},
		label: {},
		modelValue: { type: Boolean }
	},
	emits: ["update:modelValue", "copyAnchorUrl"],
	setup(__props, { emit: __emit }) {
		const emit = __emit;
		/** The trigger owns `id` as its deep link target, so the region it controls needs one of its own */
		const contentId = computed(() => `${__props.id}-content`);
		/**
		* Name the trigger after the heading it renders by pointing `aria-labelledby` at it,
		* rather than copying the text into an `aria-label`. Referencing the visible node keeps
		* the accessible name in sync with what is on screen, so it can never drift into a
		* WCAG 2.5.3 (Label in Name) failure the way a duplicated string can.
		*/
		const labelId = useId();
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("section", {
				"aria-label": __props.label,
				class: "collapsible-section"
			}, [createElementVNode("button", {
				id: __props.id,
				"aria-controls": __props.modelValue ? contentId.value : void 0,
				"aria-expanded": __props.modelValue,
				"aria-labelledby": unref(labelId),
				class: normalizeClass(["collapsible-section-trigger", { "collapsible-section-trigger-open": __props.modelValue }]),
				type: "button",
				onClick: _cache[1] || (_cache[1] = ($event) => emit("update:modelValue", !__props.modelValue))
			}, [createVNode(unref(ScalarIconCaretRight), {
				class: normalizeClass(["size-3 transition-transform duration-100", { "rotate-90": __props.modelValue }]),
				weight: "bold"
			}, null, 8, ["class"]), createVNode(unref(Anchor_default), {
				class: "collapsible-section-header",
				onCopyAnchorUrl: _cache[0] || (_cache[0] = () => emit("copyAnchorUrl"))
			}, {
				default: withCtx(() => [createElementVNode("span", {
					id: unref(labelId),
					class: "contents"
				}, [renderSlot(_ctx.$slots, "heading", {}, void 0, true)], 8, _hoisted_3)]),
				_: 3
			})], 10, _hoisted_2), __props.modelValue ? (openBlock(), createBlock(Section_default, {
				key: 0,
				id: contentId.value,
				class: "collapsible-section-content",
				label: __props.label
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, void 0, true)]),
				_: 3
			}, 8, ["id", "label"])) : createCommentVNode("", true)], 8, _hoisted_1);
		};
	}
});
//#endregion
export { CompactSection_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=CompactSection.vue.script.js.map