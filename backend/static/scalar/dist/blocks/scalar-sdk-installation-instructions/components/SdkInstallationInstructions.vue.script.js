import { useLocalization } from "../../../features/localization/use-localization.js";
import { getLanguageIcon } from "../helpers/language-icon.js";
import { getRenderableSdks } from "../helpers/renderable-sdks.js";
import { getVisibleTabCount } from "../helpers/visible-tab-count.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, nextTick, normalizeClass, onBeforeUnmount, onMounted, openBlock, ref, renderList, toDisplayString, unref, useId, watch, withCtx } from "vue";
import { getCustomClientIds } from "@scalar/blocks/code-example";
import { ScalarMarkdown } from "@scalar/components/markdown";
import { ScalarIcon } from "@scalar/components/icon";
import { ScalarCombobox } from "@scalar/components/combobox";
//#region src/blocks/scalar-sdk-installation-instructions/components/SdkInstallationInstructions.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { key: 0 };
var _hoisted_2 = ["id"];
var _hoisted_3 = { class: "client-libraries-content" };
var _hoisted_4 = ["aria-labelledby"];
var _hoisted_5 = [
	"id",
	"aria-selected",
	"tabindex",
	"onClick",
	"onKeydown"
];
var _hoisted_6 = { class: "client-libraries-text" };
var _hoisted_7 = { class: "client-libraries-text" };
var _hoisted_8 = {
	"aria-hidden": "true",
	class: "client-libraries-measure-clip"
};
var _hoisted_9 = { class: "client-libraries-text" };
var _hoisted_10 = { class: "client-libraries" };
var _hoisted_11 = { class: "client-libraries-text" };
var _hoisted_12 = ["aria-labelledby"];
var SdkInstallationInstructions_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SdkInstallationInstructions",
	props: {
		xScalarSdkInstallation: {},
		selectedClient: {},
		eventBus: {}
	},
	setup(__props) {
		const headingId = useId();
		/** Base id used to associate each tab with the shared panel for assistive tech */
		const baseId = useId();
		const panelId = `${baseId}-panel`;
		const { translate } = useLocalization();
		/** Only the SDKs that actually have something to show, with their resolved icon */
		const sdks = computed(() => getRenderableSdks(__props.xScalarSdkInstallation).map((sdk) => ({
			...sdk,
			icon: getLanguageIcon(sdk.lang)
		})));
		/** Index of the currently selected SDK */
		const selectedIndex = ref(0);
		/** The currently selected SDK */
		const selected = computed(() => sdks.value[selectedIndex.value]);
		/**
		* The `custom/<lang>` client id for each SDK, aligned by index with `sdks`.
		*
		* We reuse the exact id scheme the operation code samples use for their custom
		* examples, so selecting a language here resolves to the same id those samples
		* are keyed by — that shared id is what keeps the two surfaces in sync.
		*/
		const sdkClientIds = computed(() => getCustomClientIds(sdks.value.map((sdk) => ({
			lang: sdk.lang,
			source: ""
		}))));
		/** Select an SDK by index and broadcast it so the operation code samples follow */
		const select = (index) => {
			selectedIndex.value = index;
			const id = sdkClientIds.value[index];
			if (id) __props.eventBus?.emit("workspace:update:selected-client", id);
		};
		watch([() => __props.selectedClient, sdkClientIds], ([client, ids]) => {
			const matched = client ? ids.findIndex((id) => id === client) : -1;
			if (matched >= 0) selectedIndex.value = matched;
			else if (selectedIndex.value > ids.length - 1) selectedIndex.value = 0;
		}, { immediate: true });
		watch(() => sdks.value.map((sdk) => sdk.lang).join("\n"), () => void nextTick(measure));
		/** The full row, measured for available width (the tab strip's own width depends on the outcome) */
		const rowRef = ref();
		/** The tab strip, used to move focus between tabs */
		const tabsRef = ref();
		const measureRef = ref();
		/** The width available to render the tabs in */
		const availableWidth = ref(0);
		/** The natural width of each tab */
		const tabWidths = ref([]);
		/** The width of the "More" dropdown trigger */
		const moreWidth = ref(0);
		/** Measure the available width and the natural width of each tab */
		const measure = () => {
			const measureEl = measureRef.value;
			const rowEl = rowRef.value;
			if (!measureEl || !rowEl) return;
			const widths = Array.from(measureEl.children).map((child) => child.offsetWidth);
			tabWidths.value = widths.slice(0, sdks.value.length);
			moreWidth.value = widths[sdks.value.length] ?? 0;
			availableWidth.value = rowEl.clientWidth;
		};
		/** How many tabs fit inline before we need the "More" dropdown */
		const visibleCount = computed(() => {
			if (!tabWidths.value.length || availableWidth.value <= 0) return sdks.value.length;
			return getVisibleTabCount(tabWidths.value, availableWidth.value, moreWidth.value);
		});
		/** The SDKs shown as inline tabs */
		const visibleSdks = computed(() => sdks.value.slice(0, visibleCount.value));
		/** Whether the selected SDK lives in the "More" dropdown */
		const isMoreActive = computed(() => selectedIndex.value >= visibleCount.value);
		/** The overflowing SDKs as combobox options, keyed by their original index */
		const moreOptions = computed(() => sdks.value.slice(visibleCount.value).map((sdk, index) => ({
			id: String(visibleCount.value + index),
			label: sdk.lang
		})));
		/** The selected option within the "More" dropdown, if any */
		const selectedMoreOption = computed(() => moreOptions.value.find((option) => option.id === String(selectedIndex.value)));
		const selectMore = (option) => {
			if (option) select(Number(option.id));
		};
		/** The id of the tab that labels the panel (falls back to the heading when the selection lives in "More") */
		const activeTabId = computed(() => isMoreActive.value ? headingId : `${baseId}-tab-${selectedIndex.value}`);
		/**
		* The visible tab that holds the roving tabindex. When the selection lives in
		* the "More" dropdown, that trigger is the tab stop instead, so no inline tab
		* should be focusable.
		*/
		const tabStopIndex = computed(() => isMoreActive.value ? -1 : selectedIndex.value);
		/** Focus a visible tab by index after the DOM has settled */
		const focusTab = (index) => {
			nextTick(() => {
				tabsRef.value?.querySelectorAll("[role=\"tab\"]")[index]?.focus();
			});
		};
		/** Arrow / Home / End keyboard navigation across the visible tabs (WAI-ARIA tabs pattern) */
		const onTabKeydown = (event, index) => {
			const lastVisible = visibleCount.value - 1;
			let next = index;
			switch (event.key) {
				case "ArrowRight":
				case "ArrowDown":
					next = index >= lastVisible ? 0 : index + 1;
					break;
				case "ArrowLeft":
				case "ArrowUp":
					next = index <= 0 ? lastVisible : index - 1;
					break;
				case "Home":
					next = 0;
					break;
				case "End":
					next = lastVisible;
					break;
				default: return;
			}
			event.preventDefault();
			select(next);
			focusTab(next);
		};
		let observer;
		let frame = 0;
		/** Coalesce resize bursts into a single measure before the next paint */
		const scheduleMeasure = () => {
			if (typeof requestAnimationFrame === "undefined") {
				measure();
				return;
			}
			cancelAnimationFrame(frame);
			frame = requestAnimationFrame(measure);
		};
		onMounted(() => {
			if (typeof ResizeObserver !== "undefined") {
				observer = new ResizeObserver(scheduleMeasure);
				if (rowRef.value) observer.observe(rowRef.value);
				if (measureRef.value) observer.observe(measureRef.value);
			}
			measure();
		});
		onBeforeUnmount(() => {
			observer?.disconnect();
			if (typeof cancelAnimationFrame !== "undefined") cancelAnimationFrame(frame);
		});
		return (_ctx, _cache) => {
			return sdks.value.length ? (openBlock(), createElementBlock("div", _hoisted_1, [
				createElementVNode("div", {
					id: unref(headingId),
					class: "client-libraries-heading"
				}, toDisplayString(unref(translate)("clientLibraries.heading")), 9, _hoisted_2),
				createElementVNode("div", _hoisted_3, [createElementVNode("div", {
					ref_key: "rowRef",
					ref: rowRef,
					class: "client-libraries-row"
				}, [createElementVNode("div", {
					ref_key: "tabsRef",
					ref: tabsRef,
					"aria-labelledby": unref(headingId),
					class: "client-libraries-tabs",
					role: "tablist"
				}, [(openBlock(true), createElementBlock(Fragment, null, renderList(visibleSdks.value, (sdk, index) => {
					return openBlock(), createElementBlock("button", {
						id: `${unref(baseId)}-tab-${index}`,
						key: index,
						"aria-controls": panelId,
						"aria-selected": index === selectedIndex.value,
						class: normalizeClass(["client-libraries", { "client-libraries__active": index === selectedIndex.value }]),
						role: "tab",
						tabindex: index === tabStopIndex.value ? 0 : -1,
						type: "button",
						onClick: ($event) => select(index),
						onKeydown: ($event) => onTabKeydown($event, index)
					}, [sdk.icon ? (openBlock(), createBlock(unref(ScalarIcon), {
						key: 0,
						class: "client-libraries-icon",
						icon: sdk.icon
					}, null, 8, ["icon"])) : createCommentVNode("", true), createElementVNode("span", _hoisted_6, toDisplayString(sdk.lang), 1)], 42, _hoisted_5);
				}), 128))], 8, _hoisted_4), visibleCount.value < sdks.value.length ? (openBlock(), createBlock(unref(ScalarCombobox), {
					key: 0,
					modelValue: selectedMoreOption.value,
					options: moreOptions.value,
					placement: "bottom-end",
					teleport: "",
					"onUpdate:modelValue": selectMore
				}, {
					default: withCtx(() => [createElementVNode("button", {
						class: normalizeClass(["client-libraries client-libraries-more", { "client-libraries__active": isMoreActive.value }]),
						type: "button"
					}, [createVNode(unref(ScalarIcon), {
						class: "client-libraries-icon",
						icon: isMoreActive.value && selected.value?.icon ? selected.value.icon : "Ellipses"
					}, null, 8, ["icon"]), createElementVNode("span", _hoisted_7, toDisplayString(unref(translate)("clientLibraries.more")), 1)], 2)]),
					_: 1
				}, 8, ["modelValue", "options"])) : createCommentVNode("", true)], 512), createElementVNode("div", _hoisted_8, [createElementVNode("div", {
					ref_key: "measureRef",
					ref: measureRef,
					class: "client-libraries-row client-libraries-row--measure"
				}, [(openBlock(true), createElementBlock(Fragment, null, renderList(sdks.value, (sdk, index) => {
					return openBlock(), createElementBlock("span", {
						key: index,
						class: "client-libraries"
					}, [sdk.icon ? (openBlock(), createBlock(unref(ScalarIcon), {
						key: 0,
						class: "client-libraries-icon",
						icon: sdk.icon
					}, null, 8, ["icon"])) : createCommentVNode("", true), createElementVNode("span", _hoisted_9, toDisplayString(sdk.lang), 1)]);
				}), 128)), createElementVNode("span", _hoisted_10, [createVNode(unref(ScalarIcon), {
					class: "client-libraries-icon",
					icon: "Ellipses"
				}), createElementVNode("span", _hoisted_11, toDisplayString(unref(translate)("clientLibraries.more")), 1)])], 512)])]),
				selected.value?.description ? (openBlock(), createElementBlock("div", {
					key: 0,
					id: panelId,
					"aria-labelledby": activeTabId.value,
					class: "selected-client",
					role: "tabpanel",
					tabindex: "0"
				}, [createVNode(unref(ScalarMarkdown), { value: selected.value.description }, null, 8, ["value"])], 8, _hoisted_12)) : createCommentVNode("", true)
			])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SdkInstallationInstructions_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SdkInstallationInstructions.vue.script.js.map