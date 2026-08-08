import { useLocalization } from "../../localization/use-localization.js";
import { uploadTempDocument } from "../../../helpers/upload-temp-document.js";
import ApiReferenceToolbarBlurb_default from "./ApiReferenceToolbarBlurb.vue.js";
import { Fragment, createBlock, createElementBlock, createTextVNode, createVNode, defineComponent, mergeModels, openBlock, toDisplayString, unref, useModel, withCtx } from "vue";
import { useToasts } from "@scalar/use-toasts";
import { useLoadingState } from "@scalar/components/loading";
import { ScalarButton } from "@scalar/components/button";
import { ScalarTextInputCopy } from "@scalar/components/text-input";
//#region src/features/developer-tools/components/ApiReferenceToolbarShareTemporary.vue?vue&type=script&setup=true&lang.ts
var ApiReferenceToolbarShareTemporary_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReferenceToolbarShareTemporary",
	props: /*@__PURE__*/ mergeModels({
		workspace: {},
		externalUrls: {}
	}, {
		"url": {},
		"urlModifiers": {}
	}),
	emits: ["update:url"],
	setup(__props) {
		const { toast } = useToasts();
		const loader = useLoadingState();
		const { translate } = useLocalization();
		const tempDocUrl = useModel(__props, "url");
		async function generateTemporaryLink() {
			if (loader.isLoading || !__props.workspace || !!tempDocUrl.value) return;
			loader.start();
			const document = __props.workspace.exportActiveDocument("json");
			if (!document) {
				toast(translate("developerTools.unableToExportDocument"), "error");
				await loader.invalidate();
				return;
			}
			try {
				const url = await uploadTempDocument(document, __props.externalUrls);
				await loader.validate({
					duration: 900,
					persist: true
				});
				tempDocUrl.value = url;
			} catch (error) {
				const message = error instanceof Error ? error.message : translate("developerTools.unknownError");
				toast(message, "error");
				await loader.invalidate();
			}
		}
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [tempDocUrl.value ? (openBlock(), createBlock(unref(ScalarTextInputCopy), {
				key: 0,
				immediate: "",
				modelValue: tempDocUrl.value,
				name: "temporary-link",
				placeholder: `${__props.externalUrls.registryUrl}/share/apis/…`
			}, null, 8, ["modelValue", "placeholder"])) : (openBlock(), createBlock(unref(ScalarButton), {
				key: 1,
				class: "h-auto p-2.5",
				loader: unref(loader),
				variant: "gradient",
				onClick: generateTemporaryLink
			}, {
				default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.uploadDocument")), 1)]),
				_: 1
			}, 8, ["loader"])), createVNode(ApiReferenceToolbarBlurb_default, { class: "-mt-1" }, {
				default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("developerTools.temporaryLinkExpiration")), 1)]),
				_: 1
			})], 64);
		};
	}
});
//#endregion
export { ApiReferenceToolbarShareTemporary_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReferenceToolbarShareTemporary.vue.script.js.map