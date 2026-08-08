import { useLocalization } from "../../localization/use-localization.js";
import { uploadTempDocument } from "../../../helpers/upload-temp-document.js";
import { createBlock, createTextVNode, defineComponent, mergeModels, nextTick, openBlock, renderSlot, toDisplayString, unref, useModel, withCtx } from "vue";
import { useToasts } from "@scalar/use-toasts";
import { useLoadingState } from "@scalar/components/loading";
import { ScalarButton } from "@scalar/components/button";
//#region src/features/developer-tools/components/ApiReferenceToolbarRegisterButton.vue?vue&type=script&setup=true&lang.ts
var ApiReferenceToolbarRegisterButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarRegisterButton",
	props: /*@__PURE__*/ mergeModels({
		workspace: {},
		externalUrls: {},
		sdks: { default: () => [] }
	}, {
		"url": {},
		"urlModifiers": {}
	}),
	emits: ["update:url"],
	setup(__props) {
		const tempDocUrl = useModel(__props, "url");
		const { toast } = useToasts();
		const loader = useLoadingState();
		const { translate } = useLocalization();
		/** Open the registration link in a new tab */
		function openRegisterLink(docUrl) {
			const url = new URL(`${__props.externalUrls.dashboardUrl}/register`);
			url.searchParams.set("url", docUrl);
			__props.sdks.forEach((sdk) => url.searchParams.append("sdk", sdk));
			window.open(url.toString(), "_blank");
		}
		/** Generate and open the registration link */
		async function generateRegisterLink() {
			if (loader.isLoading || !__props.workspace) return;
			if (tempDocUrl.value) {
				openRegisterLink(tempDocUrl.value);
				return;
			}
			loader.start();
			const document = __props.workspace.exportActiveDocument("json");
			if (!document) {
				toast(translate("developerTools.unableToExportDocument"), "error");
				await loader.invalidate();
				return;
			}
			try {
				tempDocUrl.value = await uploadTempDocument(document, __props.externalUrls);
				await loader.validate();
				openRegisterLink(tempDocUrl.value);
				await nextTick();
				await loader.clear();
			} catch (error) {
				const message = error instanceof Error ? error.message : translate("developerTools.unknownError");
				toast(message, "error");
				await loader.invalidate();
			}
		}
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(ScalarButton), {
				class: "h-auto p-2.5",
				loader: unref(loader),
				onClick: generateRegisterLink
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default", {}, () => [createTextVNode(toDisplayString(unref(translate)("developerTools.generate")), 1)])]),
				_: 3
			}, 8, ["loader"]);
		};
	}
});
//#endregion
export { ApiReferenceToolbarRegisterButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarRegisterButton.vue.script.js.map