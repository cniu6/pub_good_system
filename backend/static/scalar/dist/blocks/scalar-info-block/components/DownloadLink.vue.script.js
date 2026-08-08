import { useLocalization } from "../../../features/localization/use-localization.js";
import Badge_default from "../../../components/Badge/Badge.vue.js";
import { computed, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, toDisplayString, withCtx, withModifiers } from "vue";
import { sanitizeUrl } from "@scalar/helpers/url/is-safe-url";
//#region src/blocks/scalar-info-block/components/DownloadLink.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["href"];
var DownloadLink_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "DownloadLink",
	props: {
		documentDownloadType: {},
		eventBus: {},
		documentUrl: {},
		documentType: { default: "openapi" }
	},
	setup(__props) {
		const { translate } = useLocalization();
		const label = computed(() => __props.documentType === "asyncapi" ? translate("download.asyncapi") : translate("download.openapi"));
		/**
		* The document URL can be supplied by whoever controls the rendered document, so a protocol like
		* `javascript:` would execute script on click. Drop the direct link in that case.
		*/
		const safeDocumentUrl = computed(() => sanitizeUrl(__props.documentUrl));
		const handleDownloadClick = (format) => {
			__props.eventBus.emit("ui:download:document", { format });
		};
		return (_ctx, _cache) => {
			return [
				"yaml",
				"json",
				"both"
			].includes(__props.documentDownloadType) || __props.documentDownloadType === "direct" && safeDocumentUrl.value ? (openBlock(), createElementBlock("div", {
				key: 0,
				class: normalizeClass(["download-container group", { "download-both": __props.documentDownloadType === "both" }])
			}, [
				__props.documentDownloadType === "direct" && safeDocumentUrl.value ? (openBlock(), createElementBlock("a", {
					key: 0,
					class: "download-link download-button",
					href: safeDocumentUrl.value
				}, [createElementVNode("span", null, toDisplayString(label.value), 1)], 8, _hoisted_1)) : createCommentVNode("", true),
				__props.documentDownloadType === "json" || __props.documentDownloadType === "both" ? (openBlock(), createElementBlock("button", {
					key: 1,
					class: "download-button",
					type: "button",
					onClick: _cache[0] || (_cache[0] = withModifiers(() => handleDownloadClick("json"), ["prevent"]))
				}, [createElementVNode("span", null, toDisplayString(label.value), 1), createVNode(Badge_default, { class: "extension hidden group-hover:flex" }, {
					default: withCtx(() => [..._cache[2] || (_cache[2] = [createTextVNode("json", -1)])]),
					_: 1
				})])) : createCommentVNode("", true),
				__props.documentDownloadType === "yaml" || __props.documentDownloadType === "both" ? (openBlock(), createElementBlock("button", {
					key: 2,
					class: "download-button",
					type: "button",
					onClick: _cache[1] || (_cache[1] = withModifiers(() => handleDownloadClick("yaml"), ["prevent"]))
				}, [createElementVNode("span", null, toDisplayString(label.value), 1), createVNode(Badge_default, { class: "extension hidden group-hover:flex" }, {
					default: withCtx(() => [..._cache[3] || (_cache[3] = [createTextVNode("yaml", -1)])]),
					_: 1
				})])) : createCommentVNode("", true)
			], 2)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { DownloadLink_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=DownloadLink.vue.script.js.map