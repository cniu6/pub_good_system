import { useLocalization } from "../../localization/use-localization.js";
import ScreenReader_default from "../../../components/ScreenReader.vue.js";
import { computed, createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, mergeModels, mergeProps, normalizeClass, openBlock, toDisplayString, unref, useModel, withCtx, withModifiers } from "vue";
import { ScalarListbox } from "@scalar/components/listbox";
import { ScalarIconCaretDown } from "@scalar/icons";
import { ScalarButton } from "@scalar/components/button";
import { cva } from "@scalar/use-hooks/useBindCx";
//#region src/features/Operation/components/ContentTypeSelect.vue?vue&type=script&setup=true&lang.ts
var ContentTypeSelect_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	inheritAttrs: false,
	__name: "ContentTypeSelect",
	props: /*@__PURE__*/ mergeModels({ content: {} }, {
		"modelValue": { required: true },
		"modelModifiers": {}
	}),
	emits: ["update:modelValue"],
	setup(__props) {
		const { translate } = useLocalization();
		/** The selected content type with two-way binding */
		const selectedContentType = useModel(__props, "modelValue");
		const contentTypes = computed(() => Object.keys(__props.content ?? {}));
		const selectedOption = computed({
			get: () => options.value.find((option) => option.id === selectedContentType.value),
			set: (option) => {
				if (option) selectedContentType.value = option.id;
			}
		});
		const options = computed(() => {
			return contentTypes.value.map((type) => ({
				id: type,
				label: type
			}));
		});
		const contentTypeSelect = cva({
			base: "font-normal text-c-2 bg-b-1 py-1 flex items-center gap-1 rounded-full text-xs leading-none border",
			variants: { dropdown: {
				true: "hover:text-c-1 pl-2 pr-1.5 font-medium cursor-pointer",
				false: "px-2"
			} }
		});
		return (_ctx, _cache) => {
			return contentTypes.value.length > 1 ? (openBlock(), createBlock(unref(ScalarListbox), {
				key: 0,
				modelValue: selectedOption.value,
				"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => selectedOption.value = $event),
				options: options.value,
				placement: "bottom-end",
				teleport: "",
				onClick: _cache[2] || (_cache[2] = withModifiers(() => {}, ["stop"]))
			}, {
				default: withCtx(({ open }) => [createVNode(unref(ScalarButton), mergeProps({
					class: ["h-fit", unref(contentTypeSelect)({ dropdown: true })],
					variant: "ghost"
				}, _ctx.$attrs, { onClick: _cache[0] || (_cache[0] = withModifiers(() => {}, ["stop"])) }), {
					default: withCtx(() => [
						createVNode(ScreenReader_default, null, {
							default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("operation.selectedContentType")) + ": ", 1)]),
							_: 1
						}),
						createElementVNode("span", null, toDisplayString(selectedContentType.value), 1),
						createVNode(unref(ScalarIconCaretDown), {
							class: normalizeClass(["size-2.75 transition-transform duration-100", { "rotate-180": open }]),
							weight: "bold"
						}, null, 8, ["class"])
					]),
					_: 2
				}, 1040, ["class"])]),
				_: 1
			}, 8, ["modelValue", "options"])) : (openBlock(), createElementBlock("div", mergeProps({
				key: 1,
				class: ["selected-content-type", unref(contentTypeSelect)({ dropdown: false })]
			}, _ctx.$attrs, { tabindex: "0" }), [createElementVNode("span", null, toDisplayString(selectedContentType.value), 1)], 16));
		};
	}
});
//#endregion
export { ContentTypeSelect_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ContentTypeSelect.vue.script.js.map