import { useLocalization } from "../localization/use-localization.js";
import ApiReferenceToolbarTitle_default from "./components/ApiReferenceToolbarTitle.vue.js";
import DeployApiReference_default from "./components/DeployApiReference.vue.js";
import ModifyConfiguration_default from "./components/ModifyConfiguration.vue.js";
import ShareApiReference_default from "./components/ShareApiReference.vue.js";
import { Fragment, computed, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, mergeModels, openBlock, unref, useModel } from "vue";
import { isLocalUrl } from "@scalar/helpers/url/is-local-url";
//#region src/features/developer-tools/DeveloperTools.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var _hoisted_2 = { class: "-mx-2 flex max-w-(--refs-content-max-width) flex-1 items-center" };
var _hoisted_3 = { class: "flex flex-1 items-center" };
var DeveloperTools_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "DeveloperTools",
	props: /*@__PURE__*/ mergeModels({
		workspace: {},
		configuration: {},
		externalUrls: {}
	}, {
		"overrides": {},
		"overridesModifiers": {}
	}),
	emits: ["update:overrides"],
	setup(__props) {
		const overrides = useModel(__props, "overrides");
		const { translate } = useLocalization();
		const showDeveloperTools = computed(() => {
			if (__props.configuration?.showDeveloperTools === "always") return true;
			if (__props.configuration?.showDeveloperTools === "never") return false;
			if (typeof window === "undefined") return false;
			return isLocalUrl(window.location.href);
		});
		return (_ctx, _cache) => {
			return showDeveloperTools.value ? (openBlock(), createElementBlock("header", {
				key: 0,
				"aria-label": unref(translate)("developerTools.title"),
				class: "api-reference-toolbar bg-b-1 relative z-1 flex h-10 justify-center border-b px-15"
			}, [createElementVNode("div", _hoisted_2, [
				createElementVNode("div", _hoisted_3, [createVNode(ApiReferenceToolbarTitle_default)]),
				createVNode(ModifyConfiguration_default, {
					overrides: overrides.value,
					"onUpdate:overrides": _cache[0] || (_cache[0] = ($event) => overrides.value = $event),
					configuration: __props.configuration
				}, null, 8, ["overrides", "configuration"]),
				__props.workspace ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createVNode(ShareApiReference_default, {
					externalUrls: __props.externalUrls,
					workspace: __props.workspace
				}, null, 8, ["externalUrls", "workspace"]), createVNode(DeployApiReference_default, {
					externalUrls: __props.externalUrls,
					workspace: __props.workspace
				}, null, 8, ["externalUrls", "workspace"])], 64)) : createCommentVNode("", true)
			])], 8, _hoisted_1)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { DeveloperTools_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=DeveloperTools.vue.script.js.map