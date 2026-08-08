import { useLocalization } from "../../localization/use-localization.js";
import { computed, createBlock, defineComponent, openBlock, unref, useModel } from "vue";
import { ScalarCheckboxRadioGroup } from "@scalar/components/checkbox-input";
//#region src/features/developer-tools/components/ApiReferenceToolbarConfigLayout.vue?vue&type=script&setup=true&lang.ts
var ApiReferenceToolbarConfigLayout_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarConfigLayout",
	props: {
		"modelValue": {},
		"modelModifiers": {}
	},
	emits: ["update:modelValue"],
	setup(__props) {
		const model = useModel(__props, "modelValue");
		const { translate } = useLocalization();
		const options = computed(() => [{
			label: translate("developerTools.layoutModern"),
			value: "modern"
		}, {
			label: translate("developerTools.layoutClassic"),
			value: "classic"
		}]);
		const selected = computed({
			get: () => options.value.find((option) => option.value === model.value) ?? options.value[0],
			set: (option) => model.value = option.value
		});
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarCheckboxRadioGroup), {
				modelValue: selected.value,
				"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => selected.value = $event),
				options: options.value
			}, null, 8, ["modelValue", "options"]);
		};
	}
});
//#endregion
export { ApiReferenceToolbarConfigLayout_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarConfigLayout.vue.script.js.map