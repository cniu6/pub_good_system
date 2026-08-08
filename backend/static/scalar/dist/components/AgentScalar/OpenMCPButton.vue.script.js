import { useLocalization } from "../../features/localization/use-localization.js";
import { uploadTempDocument } from "../../helpers/upload-temp-document.js";
import { createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, mergeModels, nextTick, openBlock, resolveDynamicComponent, toDisplayString, unref, useModel, withCtx } from "vue";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { useToasts } from "@scalar/use-toasts";
import { ScalarIconArrowUpRight } from "@scalar/icons";
import { useLoadingState } from "@scalar/components/loading";
import { isValidUrl } from "@scalar/helpers/url/is-valid-url";
//#region src/components/AgentScalar/OpenMCPButton.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "scalar-mcp-layer" };
var OpenMCPButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "OpenMCPButton",
	props: /*@__PURE__*/ mergeModels({
		config: {},
		externalUrls: {},
		url: {},
		workspace: {}
	}, {
		"url": {},
		"urlModifiers": {}
	}),
	emits: ["update:url"],
	setup(__props) {
		const props = __props;
		const { copyToClipboard } = useClipboard();
		const { translate } = useLocalization();
		const { toast } = useToasts();
		const loader = useLoadingState();
		const hasConfig = props.config?.name || props.config?.url;
		const encoded = btoa(JSON.stringify(props.config ?? {}));
		const cursorLink = `cursor://anysphere.cursor-deeplink/mcp/install?name=${encodeURIComponent(props.config?.name ?? "")}&config=${encoded}`;
		const vscodeLink = `vscode:mcp/install?${encodeURIComponent(JSON.stringify(props.config ?? {}))}`;
		const docUrl = useModel(__props, "url");
		/** Generate and open the registration link */
		async function generateRegisterLink() {
			if (loader.isLoading || !props.workspace) return;
			if (docUrl.value && isValidUrl(docUrl.value)) {
				openRegisterLink(docUrl.value);
				return;
			}
			loader.start();
			const document = props.workspace.exportActiveDocument("json");
			if (!document) {
				toast(translate("developerTools.unableToExportDocument"), "error");
				await loader.invalidate();
				return;
			}
			try {
				docUrl.value = await uploadTempDocument(document, props.externalUrls);
				await loader.validate();
				openRegisterLink(docUrl.value);
				await nextTick();
				await loader.clear();
			} catch (error) {
				const message = error instanceof Error ? error.message : translate("developerTools.unknownError");
				toast(message, "error");
				await loader.invalidate();
			}
		}
		/** Open the registration link in a new tab */
		function openRegisterLink(documentUrl) {
			const url = new URL(`${props.externalUrls.dashboardUrl}/register`);
			url.searchParams.set("url", documentUrl);
			url.searchParams.set("createMcp", "true");
			window.open(url.toString(), "_blank");
		}
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", _hoisted_1, [
				(openBlock(), createBlock(resolveDynamicComponent(unref(hasConfig) ? "a" : "button"), {
					class: "scalar-mcp-layer-link",
					href: unref(hasConfig) ? vscodeLink : void 0,
					target: unref(hasConfig) ? "_blank" : void 0,
					type: unref(hasConfig) ? void 0 : "button",
					onClick: _cache[0] || (_cache[0] = (e) => {
						if (!unref(hasConfig)) {
							e.preventDefault();
							generateRegisterLink();
						}
					})
				}, {
					default: withCtx(() => [
						_cache[3] || (_cache[3] = createElementVNode("svg", {
							class: "mcp-logo",
							fill: "currentColor",
							height: "800",
							viewBox: "0 0 32 32",
							width: "800",
							xmlns: "http://www.w3.org/2000/svg"
						}, [createElementVNode("path", { d: "M30.865 3.448 24.282.281a1.99 1.99 0 0 0-2.276.385L9.397 12.171 3.902 8.004a1.33 1.33 0 0 0-1.703.073L.439 9.681a1.33 1.33 0 0 0-.005 1.969L5.2 15.999.434 20.348a1.33 1.33 0 0 0 .005 1.969l1.76 1.604a1.33 1.33 0 0 0 1.703.073l5.495-4.172 12.615 11.51a1.98 1.98 0 0 0 2.271.385l6.589-3.172a1.99 1.99 0 0 0 1.13-1.802V5.248c0-.766-.443-1.469-1.135-1.802zm-6.86 19.818L14.432 16l9.573-7.266z" })], -1)),
						_cache[4] || (_cache[4] = createTextVNode(" VS Code ", -1)),
						createVNode(unref(ScalarIconArrowUpRight), { class: "mcp-nav ml-auto size-4" })
					]),
					_: 1
				}, 8, [
					"href",
					"target",
					"type"
				])),
				(openBlock(), createBlock(resolveDynamicComponent(unref(hasConfig) ? "a" : "button"), {
					class: "scalar-mcp-layer-link",
					href: unref(hasConfig) ? cursorLink : void 0,
					target: unref(hasConfig) ? "_blank" : void 0,
					type: unref(hasConfig) ? void 0 : "button",
					onClick: _cache[1] || (_cache[1] = (e) => {
						if (!unref(hasConfig)) {
							e.preventDefault();
							generateRegisterLink();
						}
					})
				}, {
					default: withCtx(() => [
						_cache[5] || (_cache[5] = createElementVNode("svg", {
							class: "mcp-logo",
							viewBox: "0 0 466.73 532.09",
							xmlns: "http://www.w3.org/2000/svg"
						}, [createElementVNode("path", {
							d: "M457.43 125.94 244.42 2.96a22.13 22.13 0 0 0-22.12 0L9.3 125.94C3.55 129.26 0 135.4 0 142.05v247.99c0 6.65 3.55 12.79 9.3 16.11l213.01 122.98a22.13 22.13 0 0 0 22.12 0l213.01-122.98c5.75-3.32 9.3-9.46 9.3-16.11V142.05c0-6.65-3.55-12.79-9.3-16.11zm-13.38 26.05L238.42 508.15c-1.39 2.4-5.06 1.42-5.06-1.36V273.58c0-4.66-2.49-8.97-6.53-11.31L24.87 145.67c-2.4-1.39-1.42-5.06 1.36-5.06h411.26c5.84 0 9.49 6.33 6.57 11.39h-.01Z",
							style: { "fill": "currentColor" }
						})], -1)),
						_cache[6] || (_cache[6] = createTextVNode(" Cursor ", -1)),
						createVNode(unref(ScalarIconArrowUpRight), { class: "mcp-nav ml-auto size-4" })
					]),
					_: 1
				}, 8, [
					"href",
					"target",
					"type"
				])),
				!unref(hasConfig) ? (openBlock(), createElementBlock("div", {
					key: 0,
					class: "scalar-mcp-layer-link",
					onClick: generateRegisterLink
				}, [
					_cache[7] || (_cache[7] = createElementVNode("svg", {
						class: "mcp-logo",
						fill: "none",
						height: "173",
						viewBox: "0 0 156 173",
						width: "156",
						xmlns: "http://www.w3.org/2000/svg"
					}, [
						createElementVNode("path", {
							d: "m6 80.912 67.882-67.883c9.373-9.372 24.569-9.372 33.941 0s9.373 24.569 0 33.942L56.558 98.236",
							stroke: "currentColor",
							"stroke-linecap": "round",
							"stroke-width": "12"
						}),
						createElementVNode("path", {
							d: "m57.265 97.529 50.558-50.558c9.373-9.373 24.569-9.373 33.942 0l.353.353c9.373 9.373 9.373 24.569 0 33.941L80.725 142.66a8 8 0 0 0 0 11.313l12.606 12.607",
							stroke: "currentColor",
							"stroke-linecap": "round",
							"stroke-width": "12"
						}),
						createElementVNode("path", {
							d: "M90.853 30 40.648 80.205c-9.372 9.372-9.372 24.568 0 33.941 9.373 9.372 24.569 9.372 33.941 0l50.205-50.205",
							stroke: "currentColor",
							"stroke-linecap": "round",
							"stroke-width": "12"
						})
					], -1)),
					createTextVNode(" " + toDisplayString(unref(translate)("mcp.generate")) + " ", 1),
					createVNode(unref(ScalarIconArrowUpRight), { class: "mcp-nav ml-auto size-4" })
				])) : (openBlock(), createElementBlock("div", {
					key: 1,
					class: "scalar-mcp-layer-link",
					onClick: _cache[2] || (_cache[2] = ($event) => unref(copyToClipboard)(__props.config?.url ?? ""))
				}, [createTextVNode(toDisplayString(unref(translate)("mcp.connect")) + " ", 1), _cache[8] || (_cache[8] = createElementVNode("svg", {
					class: "mcp-logo ml-auto",
					fill: "none",
					height: "173",
					viewBox: "0 0 156 173",
					width: "156",
					xmlns: "http://www.w3.org/2000/svg"
				}, [
					createElementVNode("path", {
						d: "m6 80.912 67.882-67.883c9.373-9.372 24.569-9.372 33.941 0s9.373 24.569 0 33.942L56.558 98.236",
						stroke: "currentColor",
						"stroke-linecap": "round",
						"stroke-width": "12"
					}),
					createElementVNode("path", {
						d: "m57.265 97.529 50.558-50.558c9.373-9.373 24.569-9.373 33.942 0l.353.353c9.373 9.373 9.373 24.569 0 33.941L80.725 142.66a8 8 0 0 0 0 11.313l12.606 12.607",
						stroke: "currentColor",
						"stroke-linecap": "round",
						"stroke-width": "12"
					}),
					createElementVNode("path", {
						d: "M90.853 30 40.648 80.205c-9.372 9.372-9.372 24.568 0 33.941 9.373 9.372 24.569 9.372 33.941 0l50.205-50.205",
						stroke: "currentColor",
						"stroke-linecap": "round",
						"stroke-width": "12"
					})
				], -1))]))
			]);
		};
	}
});
//#endregion
export { OpenMCPButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=OpenMCPButton.vue.script.js.map