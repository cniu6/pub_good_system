import { useLocalization } from "../../localization/use-localization.js";
import ApiReferenceToolbarPopover_default from "./ApiReferenceToolbarPopover.vue.js";
import ApiReferenceToolbarConfigLayout_default from "./ApiReferenceToolbarConfigLayout.vue.js";
import ApiReferenceToolbarConfigLayoutOptions_default from "./ApiReferenceToolbarConfigLayoutOptions.vue.js";
import ApiReferenceToolbarConfigTheme_default from "./ApiReferenceToolbarConfigTheme.vue.js";
import { computed, createBlock, createElementVNode, createTextVNode, createVNode, defineComponent, mergeModels, openBlock, toDisplayString, unref, useModel, withCtx } from "vue";
import { ScalarCodeBlock } from "@scalar/components/code-block";
import { prettyPrintJson } from "@scalar/helpers/json/pretty-print-json";
import { ScalarFormField, ScalarFormSection } from "@scalar/components/form";
//#region src/features/developer-tools/components/ModifyConfiguration.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "flex flex-col gap-4" };
var ModifyConfiguration_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ModifyConfiguration",
	props: /*@__PURE__*/ mergeModels({ configuration: {} }, {
		"overrides": {},
		"overridesModifiers": {}
	}),
	emits: ["update:overrides"],
	setup(__props) {
		const overrides = useModel(__props, "overrides");
		const { translate } = useLocalization();
		const snippet = computed(() => {
			return prettyPrintJson({
				...overrides.value,
				...__props.configuration,
				...overrides.value
			});
		});
		const theme = computed({
			get: () => overrides.value?.theme ?? __props.configuration?.theme ?? "default",
			set: (t) => overrides.value = {
				...overrides.value,
				theme: t
			}
		});
		const layout = computed({
			get: () => overrides.value?.layout ?? __props.configuration?.layout ?? "modern",
			set: (l) => overrides.value = {
				...overrides.value,
				layout: l
			}
		});
		return (_ctx, _cache) => {
			return openBlock(), createBlock(ApiReferenceToolbarPopover_default, { class: "w-120" }, {
				label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.configure")), 1)]),
				default: withCtx(() => [createVNode(unref(ScalarFormSection), null, {
					label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.scalarConfiguration")), 1)]),
					default: withCtx(() => [createVNode(unref(ScalarCodeBlock), {
						class: "bg-b-1.5 flex max-h-40 flex-col rounded border text-sm",
						content: snippet.value,
						lang: "json"
					}, null, 8, ["content"])]),
					_: 1
				}), createElementVNode("div", _hoisted_1, [
					createVNode(unref(ScalarFormField), null, {
						label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.theme")), 1)]),
						default: withCtx(() => [createVNode(ApiReferenceToolbarConfigTheme_default, {
							modelValue: theme.value,
							"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => theme.value = $event)
						}, null, 8, ["modelValue"])]),
						_: 1
					}),
					createVNode(unref(ScalarFormField), null, {
						label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.layout")), 1)]),
						default: withCtx(() => [createVNode(ApiReferenceToolbarConfigLayout_default, {
							modelValue: layout.value,
							"onUpdate:modelValue": _cache[1] || (_cache[1] = ($event) => layout.value = $event)
						}, null, 8, ["modelValue"])]),
						_: 1
					}),
					createVNode(unref(ScalarFormField), { is: "div" }, {
						label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.layoutOptions")), 1)]),
						default: withCtx(() => [createVNode(ApiReferenceToolbarConfigLayoutOptions_default, {
							modelValue: overrides.value,
							"onUpdate:modelValue": _cache[2] || (_cache[2] = ($event) => overrides.value = $event),
							configuration: __props.configuration
						}, null, 8, ["modelValue", "configuration"])]),
						_: 1
					})
				])]),
				_: 1
			});
		};
	}
});
//#endregion
export { ModifyConfiguration_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ModifyConfiguration.vue.script.js.map