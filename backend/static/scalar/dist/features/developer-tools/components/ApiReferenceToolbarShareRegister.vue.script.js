import { useLocalization } from "../../localization/use-localization.js";
import ApiReferenceToolbarBlurb_default from "./ApiReferenceToolbarBlurb.vue.js";
import ApiReferenceToolbarRegisterButton_default from "./ApiReferenceToolbarRegisterButton.vue.js";
import { Fragment, createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderList, resolveDynamicComponent, toDisplayString, unref, withCtx } from "vue";
import { ScalarIconBookOpen, ScalarIconBracketsCurly, ScalarIconCloud, ScalarIconFileMd, ScalarIconGitBranch, ScalarIconGlobeSimple, ScalarIconLockSimple, ScalarIconPlugsConnected, ScalarIconSparkle, ScalarIconWarningOctagon } from "@scalar/icons";
//#region src/features/developer-tools/components/ApiReferenceToolbarShareRegister.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "text-c-2 mb-2 grid grid-cols-2 gap-2.5 font-medium" };
var ApiReferenceToolbarShareRegister_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarShareRegister",
	props: {
		workspace: {},
		externalUrls: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const FEATURES = [
			{
				icon: ScalarIconLockSimple,
				labelKey: "developerTools.passwordProtection"
			},
			{
				icon: ScalarIconGlobeSimple,
				labelKey: "developerTools.customDomains"
			},
			{
				icon: ScalarIconBookOpen,
				labelKey: "developerTools.freeFormContent"
			},
			{
				icon: ScalarIconCloud,
				labelKey: "developerTools.cdnInfrastructure"
			},
			{
				icon: ScalarIconGitBranch,
				labelKey: "developerTools.pullFromGitHub"
			},
			{
				icon: ScalarIconFileMd,
				labelKey: "developerTools.markdownMdx"
			},
			{
				icon: ScalarIconWarningOctagon,
				labelKey: "developerTools.spectralLinting"
			},
			{
				icon: ScalarIconBracketsCurly,
				labelKey: "developerTools.jsonSchemaHosting"
			},
			{
				icon: ScalarIconSparkle,
				labelKey: "developerTools.askAi"
			},
			{
				icon: ScalarIconPlugsConnected,
				labelKey: "developerTools.mcpServers"
			}
		];
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [
				createElementVNode("ul", _hoisted_1, [(openBlock(), createElementBlock(Fragment, null, renderList(FEATURES, (feature) => {
					return createElementVNode("li", {
						key: feature.labelKey,
						class: "flex items-center gap-2"
					}, [(openBlock(), createBlock(resolveDynamicComponent(feature.icon), {
						class: "text-c-3 size-3.5",
						weight: "bold"
					})), createTextVNode(" " + toDisplayString(unref(translate)(feature.labelKey)), 1)]);
				}), 64))]),
				createVNode(ApiReferenceToolbarRegisterButton_default, {
					externalUrls: __props.externalUrls,
					workspace: __props.workspace
				}, {
					default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.deployOnScalar")), 1)]),
					_: 1
				}, 8, ["externalUrls", "workspace"]),
				createVNode(ApiReferenceToolbarBlurb_default, null, {
					default: withCtx(() => [
						createTextVNode(toDisplayString(unref(translate)("developerTools.deployFree")) + " ", 1),
						_cache[0] || (_cache[0] = createElementVNode("br", null, null, -1)),
						createTextVNode(" " + toDisplayString(unref(translate)("developerTools.additionalFeaturesMightRequire")) + " ", 1),
						_cache[1] || (_cache[1] = createElementVNode("span", null, [createElementVNode("a", {
							href: "https://scalar.com/products/docs/getting-started",
							rel: "noopener noreferrer",
							target: "_blank"
						}, " Scalar Pro. ")], -1))
					]),
					_: 1
				})
			], 64);
		};
	}
});
//#endregion
export { ApiReferenceToolbarShareRegister_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarShareRegister.vue.script.js.map