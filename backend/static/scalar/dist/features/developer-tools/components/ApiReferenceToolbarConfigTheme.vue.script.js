import { computed, createBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, unref, useModel, withCtx } from "vue";
import { presets, themeIds, themeLabels } from "@scalar/themes";
import { ScalarListboxCheckbox } from "@scalar/components/listbox";
import { ScalarIconCaretDown } from "@scalar/icons";
import { ScalarCombobox } from "@scalar/components/combobox";
import { ScalarFormInput } from "@scalar/components/form";
import { ScalarThemeSwatches } from "@scalar/components/theme-swatches";
//#region src/features/developer-tools/components/ApiReferenceToolbarConfigTheme.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "min-w-0 flex-1 truncate text-left" };
var _hoisted_2 = { class: "text-c-1 inline-block min-w-0 flex-1 truncate" };
var ApiReferenceToolbarConfigTheme_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarConfigTheme",
	props: {
		"modelValue": {},
		"modelModifiers": {}
	},
	emits: ["update:modelValue"],
	setup(__props) {
		const model = useModel(__props, "modelValue");
		const options = computed(() => themeIds.filter((id) => id !== "none").map((id) => ({
			id,
			label: themeLabels[id],
			css: presets[id].theme
		})));
		const selected = computed({
			get: () => {
				const theme = model.value ?? "default";
				return options.value.find((o) => o.id === theme) ?? options.value[0];
			},
			set: (option) => model.value = option.id
		});
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarCombobox), {
				modelValue: selected.value,
				"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => selected.value = $event),
				options: options.value,
				resize: ""
			}, {
				default: withCtx(({ open }) => [createVNode(unref(ScalarFormInput), null, {
					default: withCtx(() => [
						createElementVNode("div", _hoisted_1, toDisplayString(selected.value.label), 1),
						createVNode(unref(ScalarThemeSwatches), {
							class: "mr-2",
							css: selected.value.css
						}, null, 8, ["css"]),
						createVNode(unref(ScalarIconCaretDown), { class: normalizeClass(["size-3.5 transition-transform", { "rotate-180": open }]) }, null, 8, ["class"])
					]),
					_: 2
				}, 1024)]),
				option: withCtx(({ selected, option }) => [
					createVNode(unref(ScalarListboxCheckbox), { selected }, null, 8, ["selected"]),
					createElementVNode("span", _hoisted_2, toDisplayString(option.label), 1),
					createVNode(unref(ScalarThemeSwatches), { css: option.css }, null, 8, ["css"])
				]),
				_: 1
			}, 8, ["modelValue", "options"]);
		};
	}
});
//#endregion
export { ApiReferenceToolbarConfigTheme_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarConfigTheme.vue.script.js.map