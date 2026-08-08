import { computed, createElementBlock, createElementVNode, createVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
import { ScalarSidebarButton } from "@scalar/components/sidebar";
import { ScalarListbox } from "@scalar/components/listbox";
import { ScalarIconCaretUpDown } from "@scalar/icons";
//#region src/blocks/scalar-asyncapi-sidebar-filters-block/components/SidebarFilter.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "asyncapi-sidebar-filter min-w-0" };
var _hoisted_2 = { class: "text-c-1 truncate" };
/**
* SidebarFilter
*
* A compact picker inside the AsyncAPI sidebar filters section. Reused for the
* stacked protocol and server filters.
*
* `options` is expected to lead with an "All …" entry, which is also used as the
* fallback selection — so the first option doubles as the cleared state.
*/
var SidebarFilter_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SidebarFilter",
	props: {
		label: {},
		options: {},
		modelValue: {}
	},
	emits: ["update:modelValue"],
	setup(__props, { emit: __emit }) {
		const props = __props;
		const emit = __emit;
		/** Falls back to the first ("All …") option when nothing matches. */
		const selected = computed(() => props.options.find((o) => o.id === props.modelValue) ?? props.options[0]);
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("li", _hoisted_1, [createVNode(unref(ScalarListbox), {
				label: __props.label,
				modelValue: selected.value,
				options: __props.options,
				resize: "",
				teleport: "",
				"onUpdate:modelValue": _cache[0] || (_cache[0] = (e) => emit("update:modelValue", e.id))
			}, {
				default: withCtx(() => [createVNode(unref(ScalarSidebarButton), {
					is: "button",
					class: "w-full items-center text-left"
				}, {
					aside: withCtx(() => [createVNode(unref(ScalarIconCaretUpDown), {
						class: "text-c-1 ml-1 size-3 shrink-0 self-center",
						weight: "bold"
					})]),
					default: withCtx(() => [createElementVNode("span", _hoisted_2, toDisplayString(selected.value?.label), 1)]),
					_: 1
				})]),
				_: 1
			}, 8, [
				"label",
				"modelValue",
				"options"
			])]);
		};
	}
});
//#endregion
export { SidebarFilter_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SidebarFilter.vue.script.js.map