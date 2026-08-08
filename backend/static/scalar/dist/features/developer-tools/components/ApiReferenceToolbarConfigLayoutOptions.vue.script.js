import { useLocalization } from "../../localization/use-localization.js";
import { computed, createBlock, createTextVNode, createVNode, defineComponent, mergeModels, openBlock, toDisplayString, unref, useModel, withCtx } from "vue";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
import { ScalarFormInputGroup } from "@scalar/components/form";
import { ScalarToggleInput } from "@scalar/components/toggle";
//#region src/features/developer-tools/components/ApiReferenceToolbarConfigLayoutOptions.vue?vue&type=script&setup=true&lang.ts
var ApiReferenceToolbarConfigLayoutOptions_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarConfigLayoutOptions",
	props: /*@__PURE__*/ mergeModels({ configuration: {} }, {
		"modelValue": { default: () => ({}) },
		"modelModifiers": {}
	}),
	emits: ["update:modelValue"],
	setup(__props) {
		const model = useModel(__props, "modelValue");
		const { translate } = useLocalization();
		function getValue(key, defaultValue = false) {
			return model.value[key] ?? __props.configuration?.[key] ?? defaultValue;
		}
		function setValue(key, value, defaultValue = false) {
			if (value !== defaultValue) model.value = {
				...model.value,
				[key]: value
			};
			else model.value = Object.fromEntries(Object.entries(model.value).filter(([k]) => key !== k));
		}
		const modelsSectionLabel = computed(() => __props.configuration?.modelsSectionLabel ?? DEFAULT_MODELS_SECTION_LABEL);
		const expandAllModelsLabel = computed(() => translate("developerTools.expandAll", { label: modelsSectionLabel.value }));
		const hideModelsLabel = computed(() => translate("developerTools.hideModels", { label: modelsSectionLabel.value }));
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarFormInputGroup), null, {
				default: withCtx(() => [
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("showSidebar", true),
						"onUpdate:modelValue": _cache[0] || (_cache[0] = (v) => setValue("showSidebar", !!v, true))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.showSidebar")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("defaultOpenFirstTag", true),
						"onUpdate:modelValue": _cache[1] || (_cache[1] = (v) => setValue("defaultOpenFirstTag", !!v, true))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.defaultOpenFirstTag")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("defaultOpenAllTags"),
						"onUpdate:modelValue": _cache[2] || (_cache[2] = (v) => setValue("defaultOpenAllTags", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.defaultOpenAllTags")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("expandAllModelSections"),
						"onUpdate:modelValue": _cache[3] || (_cache[3] = (v) => setValue("expandAllModelSections", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(expandAllModelsLabel.value), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("expandAllResponses"),
						"onUpdate:modelValue": _cache[4] || (_cache[4] = (v) => setValue("expandAllResponses", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.expandAllResponses")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("hideClientButton"),
						"onUpdate:modelValue": _cache[5] || (_cache[5] = (v) => setValue("hideClientButton", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.hideClientButton")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("hideDarkModeToggle"),
						"onUpdate:modelValue": _cache[6] || (_cache[6] = (v) => setValue("hideDarkModeToggle", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.hideDarkModeToggle")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("hideModels"),
						"onUpdate:modelValue": _cache[7] || (_cache[7] = (v) => setValue("hideModels", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(hideModelsLabel.value), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("hideSearch"),
						"onUpdate:modelValue": _cache[8] || (_cache[8] = (v) => setValue("hideSearch", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.hideSearch")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("showOperationId"),
						"onUpdate:modelValue": _cache[9] || (_cache[9] = (v) => setValue("showOperationId", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.showOperationId")), 1)]),
						_: 1
					}, 8, ["modelValue"]),
					createVNode(unref(ScalarToggleInput), {
						modelValue: getValue("hideTestRequestButton"),
						"onUpdate:modelValue": _cache[10] || (_cache[10] = (v) => setValue("hideTestRequestButton", !!v))
					}, {
						default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.hideTestRequestButton")), 1)]),
						_: 1
					}, 8, ["modelValue"])
				]),
				_: 1
			});
		};
	}
});
//#endregion
export { ApiReferenceToolbarConfigLayoutOptions_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarConfigLayoutOptions.vue.script.js.map