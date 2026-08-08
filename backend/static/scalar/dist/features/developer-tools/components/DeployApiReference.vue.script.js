import { useLocalization } from "../../localization/use-localization.js";
import ApiReferenceToolbarPopover_default from "./ApiReferenceToolbarPopover.vue.js";
import ApiReferenceToolbarShareRegister_default from "./ApiReferenceToolbarShareRegister.vue.js";
import { createBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
import { ScalarFormSection } from "@scalar/components/form";
//#region src/features/developer-tools/components/DeployApiReference.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "text-c-2 mb-2 leading-normal" };
var DeployApiReference_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "DeployApiReference",
	props: {
		workspace: {},
		externalUrls: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(ApiReferenceToolbarPopover_default, { class: "w-120" }, {
				label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.deploy")), 1)]),
				default: withCtx(() => [createVNode(unref(ScalarFormSection), null, {
					label: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.scalarDocs")), 1)]),
					default: withCtx(() => [createElementVNode("p", _hoisted_1, toDisplayString(unref(translate)("developerTools.deployDescription")), 1), createVNode(ApiReferenceToolbarShareRegister_default, {
						externalUrls: __props.externalUrls,
						workspace: __props.workspace
					}, null, 8, ["externalUrls", "workspace"])]),
					_: 1
				})]),
				_: 1
			});
		};
	}
});
//#endregion
export { DeployApiReference_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=DeployApiReference.vue.script.js.map