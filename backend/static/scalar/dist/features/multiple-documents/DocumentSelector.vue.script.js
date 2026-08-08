import { computed, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref, withCtx } from "vue";
import { ScalarListbox } from "@scalar/components/listbox";
import { ScalarIconCaretDown } from "@scalar/icons";
//#region src/features/multiple-documents/DocumentSelector.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "document-selector px-3 pt-3"
};
var _hoisted_2 = {
	class: "group/dropdown-label text-c-2 hover:text-c-1 flex w-full cursor-pointer items-center gap-1 font-medium",
	type: "button"
};
var _hoisted_3 = { class: "overflow-hidden text-base text-ellipsis" };
var DocumentSelector_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "DocumentSelector",
	props: {
		options: {},
		modelValue: {}
	},
	emits: ["update:modelValue"],
	setup(__props, { emit: __emit }) {
		const props = __props;
		const emit = __emit;
		const formattedOptions = computed(() => props.options.map((o) => ({
			id: o.id,
			label: o.label
		})));
		const selected = computed(() => formattedOptions.value.find((o) => o.id === props.modelValue));
		return (_ctx, _cache) => {
			return __props.options.length > 1 ? (openBlock(), createElementBlock("div", _hoisted_1, [createVNode(unref(ScalarListbox), {
				modelValue: selected.value,
				options: formattedOptions.value,
				resize: "",
				"onUpdate:modelValue": _cache[0] || (_cache[0] = (e) => emit("update:modelValue", e.id))
			}, {
				default: withCtx(({ open }) => [createElementVNode("button", _hoisted_2, [createElementVNode("span", _hoisted_3, toDisplayString(selected.value?.label || "Select API"), 1), createVNode(unref(ScalarIconCaretDown), {
					class: normalizeClass(["size-3 text-current transition-transform", { "rotate-180": open }]),
					weight: "bold"
				}, null, 8, ["class"])])]),
				_: 1
			}, 8, ["modelValue", "options"])])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { DocumentSelector_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=DocumentSelector.vue.script.js.map