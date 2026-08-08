import { useLocalization } from "../../../features/localization/use-localization.js";
import { computed, createBlock, createCommentVNode, defineComponent, openBlock, unref } from "vue";
import { getAsyncApiDocumentSecurityRequirements } from "@scalar/workspace-store/channel-example";
import { getSecurityRequirements, getSelectedSecurity } from "@scalar/workspace-store/request-example";
import { getDocumentType, isAsyncApiDocument, isOpenApiDocument } from "@scalar/workspace-store/schemas/type-guards";
import { AuthSelector } from "@scalar/api-client/blocks/scalar-auth-selector-block";
//#region src/components/Content/Auth/Auth.vue?vue&type=script&setup=true&lang.ts
var Auth_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Auth",
	props: {
		options: {},
		authStore: {},
		document: {},
		eventBus: {},
		securitySchemes: {},
		selectedServer: {},
		environment: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* Document name used to scope auth selections in the store. Both OpenAPI and AsyncAPI
		* documents persist it on `x-scalar-navigation.name`.
		*/
		const documentName = computed(() => {
			if (isOpenApiDocument(__props.document) || isAsyncApiDocument(__props.document)) return __props.document["x-scalar-navigation"]?.name ?? "";
			return "";
		});
		/** Document type used to label the missing-type warning (OpenAPI vs AsyncAPI). */
		const documentType = computed(() => getDocumentType(__props.document));
		/**
		* Compute what the security requirements should be for the document.
		*
		* AsyncAPI has no root-level `security`, so document-wide auth is derived from the union of
		* every server's security requirements.
		*/
		const securityRequirements = computed(() => {
			if (isAsyncApiDocument(__props.document)) return getAsyncApiDocumentSecurityRequirements(__props.document);
			return getSecurityRequirements(isOpenApiDocument(__props.document) ? __props.document.security : void 0);
		});
		/** Grab the selected security for the document from the auth store */
		const documentSelectedSecurity = computed(() => __props.authStore.getAuthSelectedSchemas({
			type: "document",
			documentName: documentName.value
		}));
		/** The selected security keys for the document */
		const selectedSecurity = computed(() => getSelectedSecurity(documentSelectedSecurity.value, void 0, securityRequirements.value, __props.securitySchemes, __props.options.authentication?.preferredSecurityScheme));
		return (_ctx, _cache) => {
			return Object.keys(__props.securitySchemes).length ? (openBlock(), createBlock(unref(AuthSelector), {
				key: 0,
				canDeleteSchemes: false,
				createAnySecurityScheme: __props.options.authentication?.createAnySecurityScheme ?? false,
				documentType: documentType.value,
				environment: __props.environment,
				eventBus: __props.eventBus,
				isStatic: "",
				layout: "reference",
				meta: { type: "document" },
				options: { oauth2RedirectUri: __props.options.oauth2RedirectUri },
				persistAuth: __props.options.persistAuth,
				proxyUrl: __props.options.proxyUrl ?? "",
				securityRequirements: securityRequirements.value,
				securitySchemes: __props.securitySchemes,
				selectedSecurity: selectedSecurity.value,
				server: __props.selectedServer,
				title: unref(translate)("authentication.title")
			}, null, 8, [
				"createAnySecurityScheme",
				"documentType",
				"environment",
				"eventBus",
				"options",
				"persistAuth",
				"proxyUrl",
				"securityRequirements",
				"securitySchemes",
				"selectedSecurity",
				"server",
				"title"
			])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Auth_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Auth.vue.script.js.map