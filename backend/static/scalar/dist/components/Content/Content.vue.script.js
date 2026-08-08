import AsyncApiServerSelector_default from "../../blocks/scalar-asyncapi-server-selector-block/components/AsyncApiServerSelector.vue.js";
import ClientSelector_default from "../../blocks/scalar-client-selector-block/components/ClientSelector.vue.js";
import { getXKeysFromObject } from "../../features/specification-extension/helpers.js";
import InfoBlock_default from "../../blocks/scalar-info-block/components/InfoBlock.vue.js";
import IntroductionCardItem_default from "../../blocks/scalar-info-block/components/IntroductionCardItem.vue.js";
import { getRenderableSdks } from "../../blocks/scalar-sdk-installation-instructions/helpers/renderable-sdks.js";
import SdkInstallationInstructions_default from "../../blocks/scalar-sdk-installation-instructions/components/SdkInstallationInstructions.vue.js";
import ServerSelector_default from "../../blocks/scalar-server-selector-block/components/ServerSelector.vue.js";
import { firstLazyLoadComplete, scheduleInitialLoadComplete } from "../../helpers/lazy-bus.js";
import AsyncApiTraversedEntry_default from "./AsyncApi/AsyncApiTraversedEntry.vue.js";
import Auth_default from "./Auth/Auth.vue.js";
import TraversedEntry_default from "./Operations/TraversedEntry.vue.js";
import RenderPlugins_default from "../RenderPlugins/RenderPlugins.vue.js";
import SectionFlare_default from "../SectionFlare/SectionFlare.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, onMounted, openBlock, renderSlot, unref, withCtx } from "vue";
import { getAsyncApiServers, getSelectedAsyncApiServer } from "@scalar/workspace-store/channel-example";
import { getSelectedServer, getServers, mergeSecurity } from "@scalar/workspace-store/request-example";
import { getDocumentType, isAsyncApiDocument, isOpenApiDocument } from "@scalar/workspace-store/schemas/type-guards";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { mapHiddenClientsConfig } from "@scalar/api-client/modal/map-hidden-clients-config";
import { generateClientOptions } from "@scalar/blocks/code-example";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
//#region src/components/Content/Content.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "narrow-references-container" };
var _hoisted_2 = {
	key: 2,
	class: "h-dvh"
};
var Content_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Content",
	props: {
		infoSectionId: {},
		documentSlug: {},
		options: {},
		document: {},
		clientDocument: {},
		authStore: {},
		xScalarDefaultClient: {},
		xScalarDefaultExample: {},
		items: {},
		expandedItems: {},
		eventBus: {},
		environment: {},
		headingSlugGenerator: { type: Function }
	},
	setup(__props) {
		/** Generate all client options so that it can be shared between the top client picker and the operations */
		const clientOptions = computed(() => generateClientOptions(mapHiddenClientsConfig(__props.options.hiddenClients)));
		/**
		* Custom SDK installation instructions that actually have something to render.
		* Entries with only a `lang` are ignored so we can fall back to the client
		* selector instead of showing an empty card.
		*/
		const sdkInstallation = computed(() => getRenderableSdks(openApiDocument.value?.info?.["x-scalar-sdk-installation"]));
		/**
		* Narrow the (possibly AsyncAPI) documents to OpenAPI documents. api-reference
		* is OpenAPI-native, so AsyncAPI fields surface as undefined/empty.
		*/
		const openApiDocument = computed(() => isOpenApiDocument(__props.document) ? __props.document : void 0);
		const openApiClientDocument = computed(() => isOpenApiDocument(__props.clientDocument) ? __props.clientDocument : void 0);
		/** AsyncAPI narrow, used to render the (currently channel-only) AsyncAPI content tree. */
		const asyncApiDocument = computed(() => isAsyncApiDocument(__props.document) ? __props.document : void 0);
		/** AsyncAPI narrow of the client document, where server selection/variables are persisted. */
		const asyncApiClientDocument = computed(() => isAsyncApiDocument(__props.clientDocument) ? __props.clientDocument : void 0);
		const documentType = computed(() => getDocumentType(__props.document));
		const specificationVersion = computed(() => {
			if (isAsyncApiDocument(__props.document)) return __props.document["x-original-aas-version"] ?? __props.document.asyncapi;
			return openApiDocument.value?.["x-original-oas-version"];
		});
		/** Computed property to get all OpenAPI extension fields from the root document object */
		const documentExtensions = computed(() => getXKeysFromObject(__props.document));
		/** Computed property to get all OpenAPI extension fields from the document's info object */
		const infoExtensions = computed(() => getXKeysFromObject(__props.document?.info));
		/** Compute the servers for the document */
		const servers = computed(() => getServers(__props.options?.servers ?? openApiClientDocument.value?.servers, {
			baseServerUrl: __props.options?.baseServerURL,
			documentUrl: __props.clientDocument?.["x-scalar-original-source-url"]
		}));
		/** Compute the selected server for the document only (for now) */
		const selectedServer = computed(() => getSelectedServer(openApiClientDocument.value ?? null, null, null, servers.value));
		/**
		* Compute the AsyncAPI servers (all protocols, not just WebSocket) for the
		* document-level server selector.
		*/
		const asyncApiServers = computed(() => asyncApiClientDocument.value ? getAsyncApiServers(asyncApiClientDocument.value, { webSocketOnly: false }) : []);
		/** Compute the selected AsyncAPI server for the document */
		const asyncApiSelectedServer = computed(() => getSelectedAsyncApiServer(asyncApiClientDocument.value ?? null, asyncApiServers.value));
		/** Merge authentication config with the document security schemes */
		const securitySchemes = computed(() => {
			const sourceDocument = asyncApiClientDocument.value ?? openApiClientDocument.value;
			return mergeSecurity((sourceDocument?.components ? getResolvedRef(sourceDocument.components) : void 0)?.securitySchemes, __props.options.authentication?.securitySchemes, __props.authStore, sourceDocument?.["x-scalar-navigation"]?.name ?? "", __props.options.oauth2RedirectUri);
		});
		/**
		* Whether to show the document-level auth selector.
		*
		* For OpenAPI it stays tied to `hideTestRequestButton`, since the auth feeds the interactive
		* test client. AsyncAPI has no document-level test request, so its auth display is decoupled
		* from that flag and shown whenever a document is present.
		*/
		const showAuthSelector = computed(() => Boolean(__props.document) && (Boolean(asyncApiDocument.value) || !__props.options.hideTestRequestButton));
		/** Ensures firstLazyLoadComplete is set for documents with no Lazy sections (e.g. no operations/tags/models). */
		onMounted(() => {
			scheduleInitialLoadComplete();
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [createVNode(unref(SectionFlare_default)), createElementVNode("div", _hoisted_1, [
				renderSlot(_ctx.$slots, "start"),
				createVNode(unref(RenderPlugins_default), {
					documentSlug: __props.documentSlug,
					eventBus: __props.eventBus,
					options: __props.options,
					viewName: "content.start"
				}, null, 8, [
					"documentSlug",
					"eventBus",
					"options"
				]),
				createVNode(unref(InfoBlock_default), {
					id: __props.infoSectionId,
					documentDownloadType: __props.options.documentDownloadType,
					documentExtensions: documentExtensions.value,
					documentType: documentType.value,
					documentUrl: __props.document?.["x-scalar-original-source-url"],
					eventBus: __props.eventBus,
					externalDocs: openApiDocument.value?.externalDocs,
					headingSlugGenerator: __props.headingSlugGenerator,
					info: __props.document?.info,
					infoExtensions: infoExtensions.value,
					layout: __props.options.layout,
					specificationVersion: specificationVersion.value
				}, {
					selectors: withCtx(() => [
						createVNode(unref(ScalarErrorBoundary), null, {
							default: withCtx(() => [servers.value?.length ? (openBlock(), createBlock(unref(IntroductionCardItem_default), {
								key: 0,
								class: "scalar-reference-intro-server scalar-client introduction-card-item text-base leading-normal [--scalar-address-bar-height:0px]"
							}, {
								default: withCtx(() => [createVNode(unref(ServerSelector_default), {
									eventBus: __props.eventBus,
									selectedServer: selectedServer.value,
									servers: servers.value
								}, null, 8, [
									"eventBus",
									"selectedServer",
									"servers"
								])]),
								_: 1
							})) : createCommentVNode("", true)]),
							_: 1
						}),
						createVNode(unref(ScalarErrorBoundary), null, {
							default: withCtx(() => [asyncApiServers.value.length ? (openBlock(), createBlock(unref(IntroductionCardItem_default), {
								key: 0,
								class: "scalar-reference-intro-server scalar-client introduction-card-item text-base leading-normal [--scalar-address-bar-height:0px]"
							}, {
								default: withCtx(() => [createVNode(unref(AsyncApiServerSelector_default), {
									eventBus: __props.eventBus,
									selectedServer: asyncApiSelectedServer.value,
									servers: asyncApiServers.value
								}, null, 8, [
									"eventBus",
									"selectedServer",
									"servers"
								])]),
								_: 1
							})) : createCommentVNode("", true)]),
							_: 1
						}),
						createVNode(unref(ScalarErrorBoundary), null, {
							default: withCtx(() => [showAuthSelector.value ? (openBlock(), createBlock(unref(IntroductionCardItem_default), {
								key: 0,
								class: "scalar-reference-intro-auth scalar-client introduction-card-item leading-normal"
							}, {
								default: withCtx(() => [createVNode(unref(Auth_default), {
									authStore: __props.authStore,
									document: __props.clientDocument,
									environment: __props.environment,
									eventBus: __props.eventBus,
									options: __props.options,
									securitySchemes: securitySchemes.value,
									selectedServer: selectedServer.value
								}, null, 8, [
									"authStore",
									"document",
									"environment",
									"eventBus",
									"options",
									"securitySchemes",
									"selectedServer"
								])]),
								_: 1
							})) : createCommentVNode("", true)]),
							_: 1
						}),
						createVNode(unref(ScalarErrorBoundary), null, {
							default: withCtx(() => [sdkInstallation.value.length ? (openBlock(), createBlock(unref(IntroductionCardItem_default), {
								key: 0,
								class: "introduction-card-item scalar-reference-intro-clients"
							}, {
								default: withCtx(() => [createVNode(unref(SdkInstallationInstructions_default), {
									class: "introduction-card-item scalar-reference-intro-clients",
									eventBus: __props.eventBus,
									selectedClient: __props.xScalarDefaultClient,
									xScalarSdkInstallation: sdkInstallation.value
								}, null, 8, [
									"eventBus",
									"selectedClient",
									"xScalarSdkInstallation"
								])]),
								_: 1
							})) : clientOptions.value.length && !asyncApiDocument.value ? (openBlock(), createBlock(unref(IntroductionCardItem_default), {
								key: 1,
								class: "introduction-card-item scalar-reference-intro-clients"
							}, {
								default: withCtx(() => [createVNode(unref(ClientSelector_default), {
									class: "introduction-card-item scalar-reference-intro-clients",
									clientOptions: clientOptions.value,
									eventBus: __props.eventBus,
									selectedClient: __props.xScalarDefaultClient
								}, null, 8, [
									"clientOptions",
									"eventBus",
									"selectedClient"
								])]),
								_: 1
							})) : createCommentVNode("", true)]),
							_: 1
						})
					]),
					_: 1
				}, 8, [
					"id",
					"documentDownloadType",
					"documentExtensions",
					"documentType",
					"documentUrl",
					"eventBus",
					"externalDocs",
					"headingSlugGenerator",
					"info",
					"infoExtensions",
					"layout",
					"specificationVersion"
				]),
				__props.items.length && openApiDocument.value ? (openBlock(), createBlock(TraversedEntry_default, {
					key: 0,
					authStore: __props.authStore,
					clientOptions: clientOptions.value,
					document: openApiDocument.value,
					entries: __props.items,
					eventBus: __props.eventBus,
					expandedItems: __props.expandedItems,
					options: __props.options,
					securitySchemes: securitySchemes.value,
					selectedClient: __props.xScalarDefaultClient,
					selectedExample: __props.xScalarDefaultExample,
					selectedServer: selectedServer.value
				}, null, 8, [
					"authStore",
					"clientOptions",
					"document",
					"entries",
					"eventBus",
					"expandedItems",
					"options",
					"securitySchemes",
					"selectedClient",
					"selectedExample",
					"selectedServer"
				])) : __props.items.length && asyncApiDocument.value ? (openBlock(), createBlock(unref(AsyncApiTraversedEntry_default), {
					key: 1,
					document: asyncApiDocument.value,
					entries: __props.items,
					eventBus: __props.eventBus,
					expandedItems: __props.expandedItems,
					options: __props.options
				}, null, 8, [
					"document",
					"entries",
					"eventBus",
					"expandedItems",
					"options"
				])) : createCommentVNode("", true),
				createVNode(unref(RenderPlugins_default), {
					documentSlug: __props.documentSlug,
					eventBus: __props.eventBus,
					options: __props.options,
					viewName: "content.end"
				}, null, 8, [
					"documentSlug",
					"eventBus",
					"options"
				]),
				renderSlot(_ctx.$slots, "end"),
				!unref(firstLazyLoadComplete) ? (openBlock(), createElementBlock("div", _hoisted_2)) : createCommentVNode("", true)
			])], 64);
		};
	}
});
//#endregion
export { Content_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Content.vue.script.js.map