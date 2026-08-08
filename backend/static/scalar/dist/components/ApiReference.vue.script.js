import AsyncApiSidebarFilters_default from "../blocks/scalar-asyncapi-sidebar-filters-block/components/AsyncApiSidebarFilters.vue.js";
import { filterAsyncApiNavigation } from "../blocks/scalar-asyncapi-sidebar-filters-block/helpers/filter-async-api-navigation.js";
import { provideLocalization, resolveLocalization } from "../features/localization/use-localization.js";
import { AGENT_CONTEXT_SYMBOL, useAgent } from "../hooks/use-agent.js";
import AgentScalarButton_default from "./AgentScalar/AgentScalarButton.vue.js";
import AgentScalarDrawer_default from "./AgentScalar/AgentScalarDrawer.vue.js";
import OpenMCPButton_default from "./AgentScalar/OpenMCPButton.vue.js";
import ClassicHeader_default from "./ClassicHeader.vue.js";
import { useIntersection } from "../hooks/use-intersection.js";
import { createPluginManager } from "../plugins/plugin-manager.js";
import { PLUGIN_MANAGER_SYMBOL } from "../plugins/hooks/usePluginManager.js";
import { persistencePlugin } from "../plugins/persistence-plugin.js";
import { getIdFromUrl, makeUrlFromId, matchesBasePath, redirectUrl } from "../helpers/id-routing.js";
import { addToPriorityQueue, blockIntersection, intersectionEnabled, scrollToLazy } from "../helpers/lazy-bus.js";
import Content_default from "./Content/Content.vue.js";
import MobileHeader_default from "./MobileHeader.vue.js";
import DeveloperTools_default from "../features/developer-tools/DeveloperTools.vue.js";
import DocumentSelector_default from "../features/multiple-documents/DocumentSelector.vue.js";
import { isIntroductionEntry } from "../helpers/is-introduction-entry.js";
import SearchButton_default from "../features/Search/components/SearchButton.vue.js";
import { getSystemModePreference } from "../helpers/color-mode.js";
import { downloadDocument } from "../helpers/download.js";
import { loadAuthFromStorage, loadClientFromStorage } from "../helpers/load-from-perssistance.js";
import { mapConfigPlugins } from "../helpers/map-config-plugins.js";
import { mapConfigToWorkspaceStore } from "../helpers/map-config-to-workspace-store.js";
import { normalizeConfigurations } from "../helpers/normalize-configurations.js";
import { safeDeepClone } from "../helpers/safe-deep-clone.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createSlots, createTextVNode, createVNode, defineComponent, guardReactiveProps, normalizeClass, normalizeProps, onBeforeMount, onBeforeUnmount, onMounted, onServerPrefetch, openBlock, provide, ref, renderSlot, resolveDynamicComponent, toDisplayString, unref, useId, useTemplateRef, watch, withCtx } from "vue";
import { provideUseId } from "@headlessui/vue";
import { OpenApiClientButton } from "@scalar/api-client/blocks/operation-block";
import { ScalarColorModeToggleButton, ScalarColorModeToggleIcon } from "@scalar/components/color-mode-toggle";
import { addScalarClassesToHeadless } from "@scalar/components/helpers";
import { ScalarSidebarFooter, ScalarSidebarSection } from "@scalar/components/sidebar";
import { toJsonCompatible } from "@scalar/helpers/object/to-json-compatible";
import { slugify } from "@scalar/helpers/string/slugify";
import { isLocalUrl } from "@scalar/helpers/url/is-local-url";
import { apiReferenceConfigurationSchema } from "@scalar/schemas/api-reference";
import { ScalarSidebar, createSidebarState, scrollSidebarToTop } from "@scalar/sidebar";
import { getThemeStyles, hasObtrusiveScrollbars } from "@scalar/themes";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
import { useClipboard } from "@scalar/use-hooks/useClipboard";
import { useColorMode } from "@scalar/use-hooks/useColorMode";
import { ScalarToasts } from "@scalar/use-toasts";
import { coerce } from "@scalar/validation";
import { getAsyncApiServers } from "@scalar/workspace-store/channel-example";
import { createWorkspaceStore } from "@scalar/workspace-store/client";
import { createWorkspaceEventBus } from "@scalar/workspace-store/events";
import { getActiveEnvironment, getServers } from "@scalar/workspace-store/request-example";
import { isAsyncApiDocument, isOpenApiDocument } from "@scalar/workspace-store/schemas/type-guards";
import { useScrollLock } from "@vueuse/core";
import diff from "microdiff";
//#region src/components/ApiReference.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["dir", "lang"];
var _hoisted_2 = {
	key: 1,
	class: "flex gap-1.5 px-3 pt-3"
};
var _hoisted_3 = {
	class: "no-underline hover:underline",
	href: "https://www.scalar.com",
	rel: "noopener noreferrer",
	target: "_blank"
};
var _hoisted_4 = { key: 1 };
var _hoisted_5 = ["aria-label", "inert"];
var _hoisted_6 = { class: "w-64 empty:hidden" };
var _hoisted_7 = {
	key: 2,
	class: "references-footer"
};
var version = "1.64.1";
if (typeof window !== "undefined") console.info(`@scalar/api-reference@${version}`);
var ApiReference_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ApiReference",
	props: { configuration: {} },
	setup(__props, { expose: __expose }) {
		const props = __props;
		const { copyToClipboard } = useClipboard();
		/**
		* Used to inject the environment into built packages
		*
		* Primary use case is the open-in-client button
		*/
		const isDevelopment = false;
		/**
		* Whether scrollbars take up screen real estate.
		*
		* This defaults to `false` so the first client render matches the server (where
		* there is no DOM to measure). The real value is resolved in `onMounted` to
		* avoid a hydration mismatch on the root class.
		*/
		const obtrusiveScrollbars = ref(false);
		onMounted(() => {
			obtrusiveScrollbars.value = hasObtrusiveScrollbars();
		});
		const eventBus = createWorkspaceEventBus({ debug: isDevelopment });
		const isSidebarOpen = ref(false);
		/**
		* Due to a bug in headless UI, we need to set an ID here that can be shared across server/client
		* TODO remove this once the bug is fixed
		*
		* @see https://github.com/tailwindlabs/headlessui/issues/2979
		*/
		provideUseId(() => useId());
		/**
		* Configuration Handling
		*
		* We will normalize the configurations and store them in a computed property.
		* The active configuration will be associated with the active document.
		*/
		const configList = computed(() => normalizeConfigurations(props.configuration));
		const isMultiDocument = computed(() => Object.keys(configList.value).length > 1);
		/** Search for the source with a default attribute or use the first one */
		const activeSlug = ref(Object.values(configList.value).find((c) => c.default)?.slug ?? configList.value[Object.keys(configList.value)?.[0] ?? ""]?.slug ?? "");
		/**
		* On initial page load we need to determine if there is a valid document slug in the URL
		*
		* If there is we set the active slug to the document slug
		*/
		if (typeof window !== "undefined") {
			const url = new URL(window.location.href);
			const apiParam = url.searchParams.get("api");
			if (apiParam && configList.value[apiParam]) {
				activeSlug.value = apiParam;
				const newUrl = makeUrlFromId(getIdFromUrl(url, configList.value[apiParam].config.pathRouting?.basePath, apiParam), configList.value[apiParam].config.pathRouting?.basePath, isMultiDocument.value);
				if (newUrl) {
					newUrl.searchParams.delete("api");
					window.history.replaceState({}, "", newUrl.toString());
				}
			}
			const documentSlug = getIdFromUrl(url, Object.values(configList.value).map((c) => c.config.pathRouting?.basePath).find((p) => p ? matchesBasePath(url, p) : false), isMultiDocument.value ? void 0 : activeSlug.value).split("/")[0];
			if (documentSlug && configList.value[documentSlug]) activeSlug.value = documentSlug;
		}
		/** Computed document options list for the selector logic */
		const documentOptionList = computed(() => Object.values(configList.value).map((c) => ({
			label: c.title,
			id: c.slug
		})));
		/**
		* AsyncAPI sidebar filters (protocol + server).
		*
		* These mirror the document picker: stacked dropdowns at the top of the sidebar
		* that narrow the visible operations. State resets whenever the active document
		* changes so a filter never leaks across documents.
		*/
		const selectedProtocol = ref("");
		const selectedServer = ref("");
		/** The active document, narrowed to AsyncAPI (or `null` for OpenAPI documents). */
		const activeAsyncApiDocument = computed(() => {
			const document = workspaceStore.workspace.activeDocument;
			return isAsyncApiDocument(document) ? document : null;
		});
		watch(activeSlug, () => {
			selectedProtocol.value = "";
			selectedServer.value = "";
		});
		/** Configuration overrides to apply to the selected document (from the localhost toolbar) */
		const configurationOverrides = ref({});
		const withLocalizedConfigurationDefaults = (config, activeConfig) => {
			const localization = resolveLocalization(config.localization);
			const configuredModelsSectionLabel = configurationOverrides.value.modelsSectionLabel ?? (activeConfig?.modelsSectionLabel !== DEFAULT_MODELS_SECTION_LABEL ? activeConfig?.modelsSectionLabel : void 0);
			return {
				...config,
				modelsSectionLabel: configuredModelsSectionLabel ?? localization.translations.models.label ?? DEFAULT_MODELS_SECTION_LABEL
			};
		};
		/** Any dev toolbar modifications are merged with the active configuration */
		const mergedConfig = computed(() => {
			const activeConfig = configList.value[activeSlug.value]?.config;
			const merged = {
				...coerce(apiReferenceConfigurationSchema, {}),
				...activeConfig,
				...configurationOverrides.value
			};
			return withLocalizedConfigurationDefaults(merged, activeConfig);
		});
		const apiReferenceLocalization = provideLocalization(() => mergedConfig.value.localization);
		const sidebarOptions = computed(() => ({
			...mergedConfig.value,
			labels: {
				closeGroup: apiReferenceLocalization.translate("navigation.closeGroup"),
				httpMethod: apiReferenceLocalization.translate("common.httpMethod"),
				openGroup: apiReferenceLocalization.translate("navigation.openGroup")
			}
		}));
		/**
		* Locale string for the `lang` attribute. We normalize underscores to hyphens so values like
		* `es_MX` become valid BCP-47 language tags (`es-MX`).
		*/
		const documentLang = computed(() => apiReferenceLocalization.locale.value.replace("_", "-"));
		/** Convenience break out var to determine which routing mode we are using */
		const basePath = computed(() => mergedConfig.value.pathRouting?.basePath);
		const themeStyle = computed(() => getThemeStyles(mergedConfig.value.theme, { fonts: mergedConfig.value.withDefaultFonts }));
		/**
		* Custom CSS plus the theme styles, injected into a single `<style>` tag.
		*
		* This is rendered with `v-html` so the CSS is emitted verbatim. Interpolating
		* it as text content makes Vue HTML-escape characters like `"` into `&quot;` on
		* the server while the client keeps `"`, which both breaks the CSS and causes a
		* hydration mismatch.
		*
		* A closing `</style>` tag is neutralized first. It never appears in valid CSS,
		* but during server rendering the string lands in the HTML stream unescaped, so
		* `customCss` coming from a docs platform where readers can supply their own
		* theme would otherwise be able to close the tag and open a `<script>`.
		*/
		const styleContent = computed(() => `${mergedConfig.value.customCss ?? ""}\n${themeStyle.value}`.replace(/<\/style/gi, "<\\/style"));
		/** Navigation State Handling */
		const collectWebhooks = (entries) => entries.flatMap((entry) => {
			const nested = "children" in entry && entry.children ? collectWebhooks(entry.children) : [];
			return entry.type === "webhook" ? [{
				name: entry.name,
				method: entry.method,
				id: entry.id
			}, ...nested] : nested;
		});
		if (typeof window !== "undefined") {
			const canonical = redirectUrl(window.location.href, slugify(mergedConfig.value.modelsSectionLabel ?? DEFAULT_MODELS_SECTION_LABEL), activeSlug.value, isMultiDocument.value, mergedConfig.value.pathRouting?.basePath);
			if (canonical) window.history.replaceState({}, "", canonical.toString());
		}
		if (mergedConfig.value.redirect && typeof window !== "undefined") {
			const newPath = mergedConfig.value.redirect((mergedConfig.value.pathRouting ? window.location.pathname : "") + window.location.hash);
			if (newPath) window.history.replaceState({}, "", newPath);
		}
		/**
		* Sets the active slug and updates the URL with the selected document slug
		*
		* If an element ID is passed in we will configure the path or hash routing
		*/
		function syncSlugAndUrlWithDocument(slug, elementId, config) {
			const url = makeUrlFromId(elementId || slug, config.pathRouting?.basePath, isMultiDocument.value);
			if (url) window.history.replaceState({}, "", url.toString());
			activeSlug.value = slug;
		}
		/** Workspace Store Initialization */
		/**
		* Initializes the new client workspace store.
		*/
		const workspaceStore = createWorkspaceStore({ verbose: isDevelopment });
		/**
		* We need to keep the client store separate from the workspace store
		* This is because we want the client store to be a playground where users can test out their requests without affecting the references store
		*/
		const clientStore = createWorkspaceStore({
			verbose: isDevelopment,
			plugins: [persistencePlugin({ persistAuth: () => mergedConfig.value.persistAuth ?? false })]
		});
		/**
		* Plugin injection is not reactive. All plugins must be provided at first render.
		*
		* Created after the client store so the auth accessor below can read from it — plugin `onInit`
		* hooks may call `auth` synchronously during `notifyInit`. The reference-side Authentication panel
		* (Content → Auth.vue) persists credentials into `clientStore.auth`, so plugins must read the same
		* store to see what the user entered.
		*/
		const pluginManager = createPluginManager({
			plugins: Object.values(configList.value).flatMap((c) => c.config.plugins ?? []),
			/**
			* Read-only view of the global authentication state, so plugins can read stored secrets and
			* the selected security schemes without being able to mutate them. Wraps the client store's
			* auth methods (rather than passing the store directly) to keep the setters out of the plugin API.
			*
			* The getters return a deep copy (`export` already snapshots internally, the others go through
			* `toJsonCompatible`) so plugins receive plain data rather than the store's live reactive proxies —
			* mutating what they get back can never leak into the store.
			*/
			auth: {
				export: () => clientStore.auth.export(),
				getAuthSecrets: (documentName, schemeName) => toJsonCompatible(clientStore.auth.getAuthSecrets(documentName, schemeName)),
				getAuthSelectedSchemas: (payload) => toJsonCompatible(clientStore.auth.getAuthSelectedSchemas(payload))
			}
		});
		provide(PLUGIN_MANAGER_SYMBOL, pluginManager);
		pluginManager.notifyInit(mergedConfig.value);
		watch(mergedConfig, (config) => pluginManager.notifyConfigChange(config));
		const { toggleColorMode, isDarkMode } = useColorMode({
			initialColorMode: {
				true: "dark",
				false: "light",
				undefined: "system"
			}[String(mergedConfig.value.darkMode)],
			overrideColorMode: mergedConfig.value.forceDarkModeState
		});
		/**
		* The active document passed to the search modal. Both OpenAPI and AsyncAPI
		* documents are surfaced so the search index can pick up info.description
		* headings from either spec; AsyncAPI-specific entries (channels, operations,
		* messages) are not indexed yet.
		*/
		const activeSearchableDocument = computed(() => workspaceStore.workspace.activeDocument);
		/**
		* Sidebar entries contributed by plugin views (content.start / content.end).
		*
		* Each entry reuses the same id as the rendered plugin component, so the existing
		* navigation and scroll-spy logic can scroll to and highlight it. Plugins are static for
		* the lifetime of the manager, so this only needs to be resolved once.
		*/
		const pluginSidebarEntries = computed(() => pluginManager.getSidebarEntries(activeSlug.value).reduce((grouped, entry) => {
			grouped[entry.viewName].push({
				id: entry.id,
				title: entry.label,
				type: "text"
			});
			return grouped;
		}, {
			"content.start": [],
			"content.end": []
		}));
		/**
		* Localize the synthetic labels we generate for the sidebar (introduction, webhooks, and the models
		* section). Entries are only cloned when a label actually changes, so the common English-default case
		* does not allocate a new navigation tree on every recompute.
		*/
		const localizeNavigationEntries = (entries) => {
			const introductionTitle = apiReferenceLocalization.translate("navigation.introduction");
			const webhooksTitle = apiReferenceLocalization.translate("navigation.webhooks");
			const modelsSectionLabel = mergedConfig.value.modelsSectionLabel ?? DEFAULT_MODELS_SECTION_LABEL;
			const localize = (list) => {
				let changed = false;
				const result = list.map((entry) => {
					let localized = entry;
					if (isIntroductionEntry(entry) && entry.title !== introductionTitle) localized = {
						...entry,
						title: introductionTitle
					};
					else if (entry.type === "tag" && entry.isWebhooks === true && (entry.title !== webhooksTitle || entry.name !== webhooksTitle)) localized = {
						...entry,
						title: webhooksTitle,
						name: webhooksTitle
					};
					else if (entry.type === "models" && (entry.title !== modelsSectionLabel || entry.name !== modelsSectionLabel)) localized = {
						...entry,
						title: modelsSectionLabel,
						name: modelsSectionLabel
					};
					if ("children" in entry && entry.children) {
						const localizedChildren = localize(entry.children);
						if (localizedChildren !== entry.children) {
							localized = localized === entry ? { ...entry } : localized;
							localized.children = localizedChildren;
						}
					}
					if (localized !== entry) changed = true;
					return localized;
				});
				return changed ? result : list;
			};
			return localize(entries);
		};
		/** Initialize the sidebar */
		const sidebarState = createSidebarState(computed(() => {
			return Object.entries(workspaceStore.workspace.documents).map(([slug, document]) => {
				const children = document["x-scalar-navigation"]?.children ?? [];
				const childrenWithPlugins = slug === activeSlug.value ? [
					...pluginSidebarEntries.value["content.start"],
					...localizeNavigationEntries(children),
					...pluginSidebarEntries.value["content.end"]
				] : localizeNavigationEntries(children);
				return {
					id: slug,
					type: "document",
					description: document.info.description,
					name: document.info.title ?? slug,
					title: document.info.title ?? slug,
					children: childrenWithPlugins
				};
			});
		}), { hooks: {} });
		/** Recursively set all children of the given items to open */
		const setChildrenOpen = (items) => {
			items.forEach((item) => {
				if (item.type === "tag" || item.type === "models") sidebarState.setExpanded(item.id, true);
				if ("children" in item && item.children) setChildrenOpen(item.children);
			});
		};
		/** We get the sub items for the sidebar based on the configuration/document slug */
		const sidebarItems = computed(() => {
			const config = mergedConfig.value;
			if (!config) return [];
			const rawDocItems = sidebarState.items.value.find((item) => item.id === activeSlug.value)?.children ?? [];
			const docItems = activeAsyncApiDocument.value ? filterAsyncApiNavigation(rawDocItems, activeAsyncApiDocument.value, {
				protocol: selectedProtocol.value,
				server: selectedServer.value
			}) : rawDocItems;
			if (config.defaultOpenAllTags) setChildrenOpen(docItems);
			if (config.expandAllModelSections) {
				const models = docItems.find((item) => item.type === "models");
				if (models) {
					sidebarState.setExpanded(models.id, true);
					models.children?.forEach((child) => {
						sidebarState.setExpanded(child.id, true);
					});
				}
			}
			return docItems;
		});
		/** Find the sidebar entry that represents the introduction section */
		const infoSectionId = computed(() => sidebarItems.value.find(isIntroductionEntry)?.id ?? `${activeSlug.value}/description/introduction`);
		/** User for mobile navigation */
		const breadcrumb = ref("");
		const slotProps = computed(() => ({ breadcrumb: breadcrumb.value }));
		const setBreadcrumb = (id) => {
			const item = sidebarState.getEntryById(id);
			if (!item || item.type === "document") breadcrumb.value = "";
			else breadcrumb.value = item.title;
		};
		const scrollToLazyElement = (id) => {
			setBreadcrumb(id);
			sidebarState.setSelected(id);
			scrollToLazy(id, sidebarState.setExpanded, sidebarState.getEntryById);
		};
		/**
		* Updates the browser tab title via the user-provided `setPageTitle` callback.
		*
		* Called whenever the section in view changes — on sidebar clicks, on scroll, and
		* when switching documents — so the tab title always reflects what the reader sees.
		*/
		const updatePageTitle = (id) => {
			const setPageTitle = mergedConfig.value?.setPageTitle;
			const entry = sidebarState.getEntryById(id);
			if (!setPageTitle || typeof document === "undefined" || !entry?.title) return;
			const activeDocument = workspaceStore.workspace.activeDocument;
			document.title = setPageTitle({
				title: entry.title,
				document: {
					title: activeDocument?.info?.title ?? activeSlug.value,
					slug: activeSlug.value
				}
			});
		};
		/** Maps some config values to the workspace store to keep it reactive */
		mapConfigToWorkspaceStore({
			config: () => mergedConfig.value,
			store: workspaceStore,
			isDarkMode
		});
		mapConfigToWorkspaceStore({
			config: () => mergedConfig.value,
			store: clientStore,
			isDarkMode
		});
		/** Merged environment variables from workspace and document levels */
		const environment = computed(() => getActiveEnvironment(workspaceStore, workspaceStore.workspace.activeDocument ?? null).environment);
		if (typeof window !== "undefined") window.dataDumpWorkspace = () => workspaceStore;
		__expose({
			eventBus,
			workspaceStore,
			sidebarItems
		});
		/**
		* Computes a mapping from model names to their sidebar entry IDs.
		* This is used for quick lookups and navigation within the sidebar.
		*/
		const modelsIndex = computed(() => {
			return sidebarItems.value.filter((item) => item.type === "models").flatMap((item) => item.children ?? []).filter((item) => item.type === "model").reduce((acc, item) => {
				acc[item.name] = item.id;
				return acc;
			}, {});
		});
		eventBus.on("scroll-to:model-by-name", ({ name }) => {
			/** Find the model in the models index */
			const model = modelsIndex.value[name];
			if (model) scrollToLazyElement(model);
		});
		const addDocument = async (input, navigationOptions) => {
			const result = await workspaceStore.addDocument(input, navigationOptions);
			const previousDocument = clientStore.workspace.documents[input.name];
			const selectedServer = previousDocument && typeof previousDocument === "object" ? previousDocument["x-scalar-selected-server"] : void 0;
			const nextDocument = safeDeepClone(workspaceStore.exportWorkspace().documents[input.name]) ?? {
				"openapi": "3.1.0",
				"info": {
					title: "",
					version: ""
				},
				"x-scalar-original-document-hash": ""
			};
			if (typeof selectedServer === "string") Object.assign(nextDocument, { "x-scalar-selected-server": selectedServer });
			clientStore.loadWorkspace({
				auth: {},
				documents: { [input.name]: nextDocument },
				intermediateDocuments: {},
				originalDocuments: {},
				overrides: {},
				history: {},
				meta: {}
			});
			return result;
		};
		/** In-flight document loads, so a background preload and a user selection never load the same document twice */
		const documentLoadPromises = /* @__PURE__ */ new Map();
		/**
		* Load a document into the workspace store by slug, fetching URL sources or using inline content.
		*
		* This does not change the active document, so it is safe to call in the background to warm up
		* documents the user has not selected yet. Repeated calls are deduplicated and it becomes a no-op
		* once the document is loaded.
		*/
		const ensureDocumentLoaded = (slug) => {
			if (workspaceStore.workspace.documents[slug]) return Promise.resolve();
			const pending = documentLoadPromises.get(slug);
			if (pending) return pending;
			const normalized = configList.value[slug];
			if (!normalized) return Promise.resolve();
			const config = withLocalizedConfigurationDefaults({
				...normalized.config,
				...configurationOverrides.value
			}, normalized.config);
			const promise = (async () => {
				const result = await addDocument(normalized.source.url ? {
					name: slug,
					url: normalized.source.url,
					fetch: config.customFetch
				} : {
					name: slug,
					document: normalized.source.content ?? {}
				}, config);
				const document = clientStore.workspace.documents[slug];
				if (result === true && isOpenApiDocument(document) && document["x-scalar-selected-server"] === void 0) {
					const servers = getServers(normalized.config.servers ?? document.servers, {
						baseServerUrl: config.baseServerURL,
						documentUrl: normalized.source.url
					});
					if (servers.length > 0) clientStore.updateDocument(slug, "x-scalar-selected-server", servers[0].url);
				}
			})().finally(() => {
				documentLoadPromises.delete(slug);
			});
			documentLoadPromises.set(slug, promise);
			return promise;
		};
		/** Whether idle preloading has been stopped, for example when the component unmounts */
		let isPreloadStopped = false;
		/** Cancels the currently scheduled idle preload callback, if one is pending */
		let cancelScheduledPreload;
		/**
		* Stop any in-progress idle preloading. Without this, an orphaned instance (an Astro view
		* transition or a `createApiReference` remount) would keep fetching and parsing documents into
		* an abandoned store after unmount.
		*/
		const stopPreloadingDocuments = () => {
			isPreloadStopped = true;
			cancelScheduledPreload?.();
			cancelScheduledPreload = void 0;
		};
		/**
		* Warm up the documents the user has not selected yet while the browser is idle, so switching
		* between documents is instant. Runs on the client only and loads one document at a time to avoid
		* a burst of fetches and parsing work competing with the active document.
		*/
		const preloadDocumentsWhenIdle = () => {
			if (typeof window === "undefined") return;
			const pendingSlugs = Object.keys(configList.value).filter((slug) => !workspaceStore.workspace.documents[slug]);
			const scheduleIdle = (callback) => {
				if (typeof window.requestIdleCallback === "function") {
					const handle = window.requestIdleCallback(callback, { timeout: 1500 });
					cancelScheduledPreload = () => window.cancelIdleCallback(handle);
				} else {
					const handle = window.setTimeout(callback, 200);
					cancelScheduledPreload = () => window.clearTimeout(handle);
				}
			};
			const loadNext = () => {
				if (isPreloadStopped) return;
				const slug = pendingSlugs.shift();
				if (!slug) return;
				ensureDocumentLoaded(slug).finally(() => {
					if (!isPreloadStopped) scheduleIdle(loadNext);
				});
			};
			scheduleIdle(loadNext);
		};
		/**
		* Handle changing the active document
		*
		* 1. If the document has not be loaded to the workspace store we set it to empty and asynchronously load it
		* 2. If the document has been loaded to the workspace store we just set it to active
		* 3. If the content from the configuration has changes we need to update the document in the workspace store
		*/
		const changeSelectedDocument = async (slug, elementId) => {
			const normalized = configList.value[slug];
			if (!normalized) {
				console.warn(`Document ${slug} not found in configList`);
				return;
			}
			const config = withLocalizedConfigurationDefaults({
				...normalized.config,
				...configurationOverrides.value
			}, normalized.config);
			const onDocumentSelectPromise = config.onDocumentSelect?.();
			syncSlugAndUrlWithDocument(slug, elementId, config);
			apiClient.value?.route({
				documentSlug: slug,
				method: "get",
				path: "/"
			});
			await ensureDocumentLoaded(slug);
			workspaceStore.update("x-scalar-active-document", slug);
			clientStore.update("x-scalar-active-document", slug);
			if (elementId && typeof window !== "undefined") {
				const canonical = redirectUrl(window.location.href, slugify(config.modelsSectionLabel ?? DEFAULT_MODELS_SECTION_LABEL), slug, isMultiDocument.value, config.pathRouting?.basePath, collectWebhooks(workspaceStore.workspace.activeDocument?.["x-scalar-navigation"]?.children ?? []));
				if (canonical) {
					window.history.replaceState({}, "", canonical.toString());
					elementId = getIdFromUrl(canonical.href, config.pathRouting?.basePath, isMultiDocument.value ? void 0 : slug) || elementId;
				}
			}
			if (config.persistAuth) loadAuthFromStorage(clientStore, slug);
			(async () => {
				await onDocumentSelectPromise;
				config.onLoaded?.(slug);
			})();
			if (elementId && elementId !== slug) scrollToLazyElement(elementId);
			else if (config.defaultOpenFirstTag) {
				const firstTag = sidebarItems.value.find((item) => item.type === "tag");
				if (firstTag) sidebarState.setExpanded(firstTag.id, true);
			}
			updatePageTitle(elementId && elementId !== slug ? elementId : slug);
		};
		/**
		* TODO:Move this to a dedicated updateDocument function in the future and
		* away from vue-reactivity based updates
		*/
		watch(() => Object.values(configList.value), async (newConfigList, oldConfigList) => {
			/**
			* Handles replacing and updating documents within the workspace store
			* when we detect configuration changes.
			*/
			const updateSource = async (updated, previous) => {
				const config = withLocalizedConfigurationDefaults({
					...updated.config,
					...configurationOverrides.value
				}, updated.config);
				/**
				* A background preload may still be loading this document against the previous
				* configuration. Wait for it to finish so the update below rebases onto the loaded
				* document instead of being skipped, which would otherwise leave stale content in the store.
				*/
				const pendingLoad = documentLoadPromises.get(updated.slug);
				if (pendingLoad) await pendingLoad;
				/** If we have not loaded the document previously we don't need to handle any updates to store */
				if (!workspaceStore.workspace.documents[updated.slug]) return;
				/** If the URL has changed we fetch and rebase */
				if (updated.source.url && updated.source.url !== previous?.source.url) {
					await addDocument({
						name: updated.slug,
						url: updated.source.url,
						fetch: config.customFetch
					}, config);
					return;
				}
				if (!updated.source.content) return;
				/**
				* We need to deeply check for document changes. Parse documents and then only rebase
				* if we detect deep changes in the two sources
				*/
				if (diff(updated.source.content, previous && "content" in previous.source ? previous.source.content ?? {} : {}).length) await addDocument({
					name: updated.slug,
					document: updated.source.content
				}, config);
			};
			newConfigList.forEach((newConfig, index) => updateSource(newConfig, oldConfigList[index]));
			const newSlugs = newConfigList.map((c) => c.slug);
			const oldSlugs = oldConfigList.map((c) => c.slug);
			if (newSlugs.length !== oldSlugs.length || !newSlugs.every((slug, index) => slug === oldSlugs[index])) await changeSelectedDocument(newSlugs[0] ?? "");
		}, { deep: true });
		/** Preload the first document during SSR */
		onServerPrefetch(() => changeSelectedDocument(activeSlug.value));
		/** Load the first document on page load */
		onBeforeMount(async () => {
			loadClientFromStorage(clientStore);
			await changeSelectedDocument(activeSlug.value, getIdFromUrl(window.location.href, configList.value[activeSlug.value]?.config.pathRouting?.basePath, isMultiDocument.value ? void 0 : activeSlug.value));
			preloadDocumentsWhenIdle();
		});
		const documentUrl = computed(() => {
			return configList.value[activeSlug.value]?.source?.url;
		});
		/**
		* Determines if Agent Scalar should be enabled based on the configuration and the current URL
		*
		* - If the agent is disabled in the configuration, it should not be enabled
		* - If the current URL is a local URL, it should be enabled
		* - If the agent key is set, it should be enabled
		*/
		const agent = useAgent({ agentEnabled: computed(() => {
			if (configList.value[activeSlug.value]?.agent?.disabled) return false;
			if (typeof window !== "undefined" && isLocalUrl(window.location.href)) return true;
			return Boolean(configList.value[activeSlug.value]?.agent?.key);
		}) });
		provide(AGENT_CONTEXT_SYMBOL, agent);
		const modal = useTemplateRef("modal");
		const apiClient = ref(null);
		onMounted(async () => {
			if (!modal.value) return;
			const { createApiClientModal } = await import("@scalar/api-client/modal");
			if (!modal.value) return;
			apiClient.value = createApiClientModal({
				el: modal.value,
				eventBus,
				workspaceStore: clientStore,
				options: mergedConfig,
				plugins: [...pluginManager.getApiClientPlugins(), ...mapConfigPlugins(mergedConfig, environment)]
			});
		});
		onBeforeUnmount(() => {
			stopPreloadingDocuments();
			pluginManager.notifyDestroy();
			apiClient.value?.app.unmount();
		});
		/** Ensure we call the onServerChange callback */
		eventBus.on("server:update:selected", ({ url }) => mergedConfig.value.onServerChange?.(url));
		/**
		* AsyncAPI servers are keyed by name, so resolve the selected name to its
		* constructed connection URL before firing onServerChange, keeping the callback
		* payload consistent with OpenAPI (a URL string).
		*/
		eventBus.on("asyncapi-server:update:selected", ({ name }) => {
			const document = clientStore.workspace.activeDocument;
			if (!isAsyncApiDocument(document)) return;
			const server = getAsyncApiServers(document, { webSocketOnly: false }).find((s) => s.name === name);
			mergedConfig.value.onServerChange?.(server?.url ?? name);
		});
		/** Download the document from the store */
		eventBus.on("ui:download:document", ({ format }) => {
			const document = workspaceStore.exportActiveDocument(format);
			if (!document) {
				console.error("No document found to download");
				return;
			}
			downloadDocument(document, activeSlug.value ?? "openapi", format);
		});
		/**
		* Handler for a direct navigation event such as a sidebar or search click
		*
		* Depending on the item type we handle a selection event differently:
		*
		* - Tag: If a tag is closed we open it and all its parents and scroll to it
		*        If a tag is open we just close the tag
		* - Operation:
		*        Open all parents and scroll to the operation
		*/
		const handleSelectSidebarEntry = (id, caller) => {
			const item = sidebarState.getEntryById(id);
			updatePageTitle(id);
			if ((item?.type === "tag" || item?.type === "models" || item?.type === "text") && sidebarState.isExpanded(id) && sidebarState.selectedItem.value === id) {
				const unblock = blockIntersection();
				sidebarState.setExpanded(id, false);
				unblock();
				return;
			}
			if (item?.type !== "tag" && item?.type !== "models") isSidebarOpen.value = false;
			scrollToLazyElement(id);
			const url = makeUrlFromId(id, basePath.value, isMultiDocument.value);
			if (url) {
				window.history.pushState({}, "", url);
				if (caller === "sidebar") mergedConfig.value.onSidebarClick?.(url.toString());
			}
			if (agent.showAgent.value) agent.closeAgent();
		};
		/** Handle a navigation item selection event */
		eventBus.on("select:nav-item", ({ id }) => handleSelectSidebarEntry(id));
		/** Handle a scroll to navigation item event */
		eventBus.on("scroll-to:nav-item", ({ id }) => handleSelectSidebarEntry(id));
		/** Handle an intersecting navigation item event */
		eventBus.on("intersecting:nav-item", ({ id }) => {
			if (!intersectionEnabled.value) return;
			sidebarState.setSelected(id);
			setBreadcrumb(id);
			updatePageTitle(id);
			scrollSidebarToTop(id);
			const url = makeUrlFromId(id, basePath.value, isMultiDocument.value);
			if (url && workspaceStore.workspace.activeDocument) window.history.replaceState({}, "", url.toString());
		});
		eventBus.on("toggle:nav-item", ({ id, open }) => {
			if (open) {
				mergedConfig.value.onShowMore?.(id);
				const entry = sidebarState.getEntryById(id);
				if (entry && "children" in entry && entry.children) {
					const first = entry.children[0];
					if (first) addToPriorityQueue(first.id);
				}
			}
			sidebarState.setExpanded(id, open ?? !sidebarState.isExpanded(id));
		});
		eventBus.on("copy-url:nav-item", ({ id }) => {
			const url = makeUrlFromId(id, basePath.value, isMultiDocument.value)?.toString();
			return url && copyToClipboard(url);
		});
		onBeforeMount(() => {
			window.history.scrollRestoration = "manual";
			addScalarClassesToHeadless();
			window.addEventListener("popstate", () => {
				const id = getIdFromUrl(window.location.href, mergedConfig.value.pathRouting?.basePath, isMultiDocument.value ? void 0 : activeSlug.value);
				if (id) scrollToLazyElement(id);
			});
		});
		const documentStartRef = useTemplateRef("documentStartRef");
		/**
		* Uses `immediate` so the sentinel fires as soon as it enters the viewport (not just at the center strip).
		* When the user scrolls away from the top, both this observer and the first section's center-strip
		* observer are intersecting simultaneously, so the section observer does not re-fire on its own.
		* The `onExit` callback bridges that gap by finding whichever section is at the viewport center
		* and re-emitting the nav event for it.
		*
		* We emit the Introduction entry rather than the document slug so this sentinel and the Introduction
		* section's own intersection observer resolve to the same entry. Otherwise the two race at the top of
		* the document and the tab title flickers between the section title and the document title.
		*/
		useIntersection(documentStartRef, () => {
			eventBus.emit("intersecting:nav-item", { id: infoSectionId.value });
		}, {
			onExit: () => {
				const centerY = window.innerHeight / 2;
				const section = document.elementsFromPoint(window.innerWidth / 2, centerY).find((el) => el.tagName === "SECTION" && el.id);
				if (section?.id) eventBus.emit("intersecting:nav-item", { id: section.id });
			},
			immediate: true
		});
		const colorMode = computed(() => {
			const mode = workspaceStore.workspace["x-scalar-color-mode"];
			if (mode === "system") return getSystemModePreference();
			return mode;
		});
		const bodyScrollLocked = useScrollLock(typeof document !== "undefined" ? document.body : null);
		watch(agent.showAgent, () => bodyScrollLocked.value = agent.showAgent.value);
		const showMCPButton = computed(() => {
			if (mergedConfig.value.mcp?.disabled) return false;
			if (typeof window !== "undefined" && isLocalUrl(window.location.href)) return true;
			if (mergedConfig.value.mcp) return true;
			return false;
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", null, [
				(openBlock(), createBlock(resolveDynamicComponent("style"), { innerHTML: styleContent.value }, null, 8, ["innerHTML"])),
				createElementVNode("div", {
					ref: "documentEl",
					class: normalizeClass(["scalar-app scalar-api-reference references-layout", [{
						"scalar-api-references-standalone-mobile": mergedConfig.value.showSidebar,
						"scalar-scrollbars-obtrusive": obtrusiveScrollbars.value,
						"references-editable": mergedConfig.value.isEditable,
						"references-sidebar": mergedConfig.value.showSidebar,
						"references-sidebar-mobile-open": isSidebarOpen.value,
						"references-classic": mergedConfig.value.layout === "classic"
					}, _ctx.$attrs.class]]),
					dir: unref(apiReferenceLocalization).direction.value,
					lang: documentLang.value
				}, [
					unref(agent).agentEnabled.value ? (openBlock(), createBlock(unref(AgentScalarDrawer_default), {
						key: 0,
						agentScalarConfiguration: configList.value[activeSlug.value]?.agent,
						externalUrls: mergedConfig.value.externalUrls,
						workspaceStore: unref(workspaceStore)
					}, null, 8, [
						"agentScalarConfiguration",
						"externalUrls",
						"workspaceStore"
					])) : createCommentVNode("", true),
					mergedConfig.value.layout === "modern" ? (openBlock(), createBlock(MobileHeader_default, {
						key: 1,
						breadcrumb: breadcrumb.value,
						isSidebarOpen: isSidebarOpen.value,
						showSidebar: mergedConfig.value.showSidebar,
						onToggleSidebar: _cache[5] || (_cache[5] = () => isSidebarOpen.value = !isSidebarOpen.value)
					}, {
						search: withCtx(() => [!mergedConfig.value.hideSearch ? (openBlock(), createBlock(SearchButton_default, {
							key: 0,
							class: "my-2",
							document: activeSearchableDocument.value,
							eventBus: unref(eventBus),
							hideModels: mergedConfig.value.hideModels,
							modelsSectionLabel: mergedConfig.value.modelsSectionLabel,
							searchHotKey: mergedConfig.value.searchHotKey,
							showSidebar: mergedConfig.value.showSidebar
						}, null, 8, [
							"document",
							"eventBus",
							"hideModels",
							"modelsSectionLabel",
							"searchHotKey",
							"showSidebar"
						])) : createCommentVNode("", true)]),
						sidebar: withCtx(({ sidebarClasses }) => [mergedConfig.value.showSidebar && mergedConfig.value.layout === "modern" ? (openBlock(), createBlock(unref(ScalarSidebar), {
							key: 0,
							"aria-label": unref(apiReferenceLocalization).translate("navigation.sidebarFor", { name: unref(workspaceStore).workspace.activeDocument?.info?.title ?? "" }),
							class: normalizeClass(["t-doc__sidebar", sidebarClasses]),
							isExpanded: unref(sidebarState).isExpanded,
							isSelected: unref(sidebarState).isSelected,
							items: sidebarItems.value,
							layout: "reference",
							options: sidebarOptions.value,
							onSelectItem: _cache[3] || (_cache[3] = (id) => handleSelectSidebarEntry(id, "sidebar")),
							onToggleGroup: _cache[4] || (_cache[4] = (id) => unref(sidebarState).setExpanded(id, !unref(sidebarState).isExpanded(id)))
						}, {
							header: withCtx(() => [
								documentOptionList.value.length > 1 ? (openBlock(), createBlock(DocumentSelector_default, {
									key: 0,
									modelValue: activeSlug.value,
									options: documentOptionList.value,
									"onUpdate:modelValue": changeSelectedDocument
								}, null, 8, ["modelValue", "options"])) : createCommentVNode("", true),
								!mergedConfig.value.hideSearch ? (openBlock(), createElementBlock("div", _hoisted_2, [createVNode(SearchButton_default, {
									document: activeSearchableDocument.value,
									eventBus: unref(eventBus),
									hideModels: mergedConfig.value.hideModels,
									modelsSectionLabel: mergedConfig.value.modelsSectionLabel,
									searchHotKey: mergedConfig.value.searchHotKey
								}, null, 8, [
									"document",
									"eventBus",
									"hideModels",
									"modelsSectionLabel",
									"searchHotKey"
								]), unref(agent).agentEnabled.value ? (openBlock(), createBlock(unref(AgentScalarButton_default), { key: 0 })) : createCommentVNode("", true)])) : createCommentVNode("", true),
								renderSlot(_ctx.$slots, "sidebar-start", normalizeProps(guardReactiveProps(slotProps.value)), void 0, true)
							]),
							before: withCtx(() => [createVNode(unref(AsyncApiSidebarFilters_default), {
								protocol: selectedProtocol.value,
								"onUpdate:protocol": _cache[0] || (_cache[0] = ($event) => selectedProtocol.value = $event),
								server: selectedServer.value,
								"onUpdate:server": _cache[1] || (_cache[1] = ($event) => selectedServer.value = $event),
								document: activeAsyncApiDocument.value
							}, null, 8, [
								"protocol",
								"server",
								"document"
							]), activeAsyncApiDocument.value ? (openBlock(), createBlock(unref(ScalarSidebarSection), {
								key: 0,
								class: "asyncapi-sidebar-document-section"
							}, {
								default: withCtx(() => [..._cache[8] || (_cache[8] = [createTextVNode(" Document ", -1)])]),
								_: 1
							})) : createCommentVNode("", true)]),
							footer: withCtx(() => [renderSlot(_ctx.$slots, "sidebar-end", normalizeProps(guardReactiveProps(slotProps.value)), () => [createVNode(unref(ScalarSidebarFooter), { class: "darklight-reference" }, {
								description: withCtx(() => [createElementVNode("a", _hoisted_3, toDisplayString(unref(apiReferenceLocalization).translate("footer.poweredByScalar")), 1)]),
								toggle: withCtx(() => [!mergedConfig.value.hideDarkModeToggle && !mergedConfig.value.forceDarkModeState ? (openBlock(), createBlock(unref(ScalarColorModeToggleButton), {
									key: 0,
									modelValue: colorMode.value === "dark",
									"onUpdate:modelValue": _cache[2] || (_cache[2] = () => unref(toggleColorMode)())
								}, null, 8, ["modelValue"])) : (openBlock(), createElementBlock("span", _hoisted_4))]),
								default: withCtx(() => [!mergedConfig.value.hideClientButton && !showMCPButton.value ? (openBlock(), createBlock(unref(OpenApiClientButton), {
									key: 0,
									buttonSource: "sidebar",
									integration: mergedConfig.value._integration,
									isDevelopment: unref(isDevelopment),
									url: documentUrl.value
								}, null, 8, [
									"integration",
									"isDevelopment",
									"url"
								])) : createCommentVNode("", true), showMCPButton.value ? (openBlock(), createBlock(unref(OpenMCPButton_default), {
									key: 1,
									config: mergedConfig.value.mcp,
									externalUrls: mergedConfig.value.externalUrls,
									isDevelopment: unref(isDevelopment),
									url: documentUrl.value,
									workspace: unref(workspaceStore)
								}, null, 8, [
									"config",
									"externalUrls",
									"isDevelopment",
									"url",
									"workspace"
								])) : createCommentVNode("", true)]),
								_: 1
							})], true)]),
							_: 3
						}, 8, [
							"aria-label",
							"class",
							"isExpanded",
							"isSelected",
							"items",
							"options"
						])) : createCommentVNode("", true)]),
						_: 3
					}, 8, [
						"breadcrumb",
						"isSidebarOpen",
						"showSidebar"
					])) : createCommentVNode("", true),
					createElementVNode("main", {
						"aria-label": unref(apiReferenceLocalization).translate("navigation.mainContent", { name: unref(workspaceStore).workspace.activeDocument?.info?.title ?? "" }),
						class: "references-rendered",
						inert: unref(agent).showAgent.value
					}, [createVNode(Content_default, {
						authStore: unref(clientStore).auth,
						clientDocument: unref(clientStore).workspace.activeDocument,
						document: unref(workspaceStore).workspace.activeDocument,
						documentSlug: activeSlug.value,
						environment: environment.value,
						eventBus: unref(eventBus),
						expandedItems: unref(sidebarState).expandedItems.value,
						headingSlugGenerator: mergedConfig.value.generateHeadingSlug ?? ((heading) => `${activeSlug.value}/description/${heading.slug}`),
						infoSectionId: infoSectionId.value,
						items: sidebarItems.value,
						options: mergedConfig.value,
						xScalarDefaultClient: unref(clientStore).workspace["x-scalar-default-client"],
						xScalarDefaultExample: unref(clientStore).workspace["x-scalar-default-example"]
					}, createSlots({
						start: withCtx(() => [
							createElementVNode("div", {
								ref_key: "documentStartRef",
								ref: documentStartRef
							}, null, 512),
							unref(workspaceStore).workspace.activeDocument ? (openBlock(), createBlock(unref(DeveloperTools_default), {
								key: 0,
								overrides: configurationOverrides.value,
								"onUpdate:overrides": _cache[6] || (_cache[6] = ($event) => configurationOverrides.value = $event),
								class: "references-developer-tools",
								configuration: mergedConfig.value,
								externalUrls: mergedConfig.value.externalUrls,
								workspace: unref(workspaceStore)
							}, null, 8, [
								"overrides",
								"configuration",
								"externalUrls",
								"workspace"
							])) : createCommentVNode("", true),
							mergedConfig.value.layout === "classic" ? (openBlock(), createBlock(ClassicHeader_default, { key: 1 }, {
								"dark-mode-toggle": withCtx(() => [!mergedConfig.value.hideDarkModeToggle && !mergedConfig.value.forceDarkModeState ? (openBlock(), createBlock(unref(ScalarColorModeToggleIcon), {
									key: 0,
									class: "text-c-2 hover:text-c-1",
									mode: colorMode.value,
									style: { "transform": "scale(1.4)" },
									variant: "icon",
									onClick: _cache[7] || (_cache[7] = () => unref(toggleColorMode)())
								}, null, 8, ["mode"])) : createCommentVNode("", true)]),
								default: withCtx(() => [createElementVNode("div", _hoisted_6, [documentOptionList.value.length > 1 ? (openBlock(), createBlock(DocumentSelector_default, {
									key: 0,
									modelValue: activeSlug.value,
									options: documentOptionList.value,
									"onUpdate:modelValue": changeSelectedDocument
								}, null, 8, ["modelValue", "options"])) : createCommentVNode("", true)]), !mergedConfig.value.hideSearch ? (openBlock(), createBlock(SearchButton_default, {
									key: 0,
									class: "t-doc__sidebar max-w-64",
									document: activeSearchableDocument.value,
									eventBus: unref(eventBus),
									hideModels: mergedConfig.value.hideModels,
									modelsSectionLabel: mergedConfig.value.modelsSectionLabel,
									searchHotKey: mergedConfig.value.searchHotKey
								}, null, 8, [
									"document",
									"eventBus",
									"hideModels",
									"modelsSectionLabel",
									"searchHotKey"
								])) : createCommentVNode("", true)]),
								_: 1
							})) : createCommentVNode("", true),
							renderSlot(_ctx.$slots, "content-start", normalizeProps(guardReactiveProps(slotProps.value)), void 0, true)
						]),
						end: withCtx(() => [renderSlot(_ctx.$slots, "content-end", normalizeProps(guardReactiveProps(slotProps.value)), void 0, true)]),
						_: 2
					}, [mergedConfig.value.isEditable ? {
						name: "empty-state",
						fn: withCtx(() => [renderSlot(_ctx.$slots, "editor-placeholder", normalizeProps(guardReactiveProps(slotProps.value)), void 0, true)]),
						key: "0"
					} : void 0]), 1032, [
						"authStore",
						"clientDocument",
						"document",
						"documentSlug",
						"environment",
						"eventBus",
						"expandedItems",
						"headingSlugGenerator",
						"infoSectionId",
						"items",
						"options",
						"xScalarDefaultClient",
						"xScalarDefaultExample"
					])], 8, _hoisted_5),
					_ctx.$slots.footer ? (openBlock(), createElementBlock("div", _hoisted_7, [renderSlot(_ctx.$slots, "footer", normalizeProps(guardReactiveProps(slotProps.value)), void 0, true)])) : createCommentVNode("", true),
					createElementVNode("div", {
						ref_key: "modal",
						ref: modal
					}, null, 512)
				], 10, _hoisted_1),
				createVNode(unref(ScalarToasts))
			]);
		};
	}
});
//#endregion
export { ApiReference_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ApiReference.vue.script.js.map