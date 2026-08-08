import { useLocalization } from "../../localization/use-localization.js";
import SearchModal_default from "./SearchModal.vue.js";
import { Fragment, createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, nextTick, normalizeClass, onBeforeUnmount, onMounted, openBlock, ref, toDisplayString, unref, watch, withCtx } from "vue";
import { ScalarSidebarSearchButton } from "@scalar/components/sidebar";
import { DEFAULT_MODELS_SECTION_LABEL } from "@scalar/types/api-reference";
import { ScalarIconMagnifyingGlass } from "@scalar/icons";
import { ScalarIconButton } from "@scalar/components/icon-button";
import { useModal } from "@scalar/components/modal";
import { isMacOS } from "@scalar/helpers/general/is-mac-os";
//#region src/features/Search/components/SearchButton.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "sr-only" };
var _hoisted_2 = {
	"aria-hidden": "true",
	class: "sidebar-search-placeholder"
};
var _hoisted_3 = { class: "sr-only" };
var _hoisted_4 = { class: "sr-only" };
var SearchButton_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SearchButton",
	props: {
		forceIcon: { type: Boolean },
		searchHotKey: { default: "k" },
		hideModels: { type: Boolean },
		modelsSectionLabel: { default: () => DEFAULT_MODELS_SECTION_LABEL },
		document: {},
		eventBus: {}
	},
	setup(__props) {
		const button = ref();
		const modalState = useModal();
		const { translate } = useLocalization();
		/**
		* Whether the user is on macOS, used to show the correct shortcut symbol.
		*
		* This must default to `false` so the server-rendered markup and the first
		* client render agree. Detecting the platform relies on `navigator`, which is
		* unavailable during SSR, so we resolve it after mount to avoid a hydration
		* mismatch.
		*/
		const isMac = ref(false);
		onMounted(() => {
			isMac.value = isMacOS();
		});
		const handleHotKey = (e) => {
			if ((isMacOS() ? e.metaKey : e.ctrlKey) && e.key === __props.searchHotKey) {
				e.preventDefault();
				e.stopPropagation();
				if (modalState.open) modalState.hide();
				else modalState.show();
			}
		};
		watch(() => modalState.open, async (next, prev) => {
			if (!next && prev) {
				await nextTick();
				button.value?.$el.focus();
			}
		});
		onMounted(() => window.addEventListener("keydown", handleHotKey));
		onBeforeUnmount(() => window.removeEventListener("keydown", handleHotKey));
		function handleClick() {
			modalState.show();
		}
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [__props.forceIcon ? (openBlock(), createBlock(unref(ScalarIconButton), {
				key: 0,
				icon: unref(ScalarIconMagnifyingGlass),
				label: unref(translate)("search.label"),
				onClick: handleClick
			}, null, 8, ["icon", "label"])) : (openBlock(), createBlock(unref(ScalarSidebarSearchButton), {
				key: 1,
				ref_key: "button",
				ref: button,
				class: normalizeClass(["w-full", _ctx.$attrs.class]),
				shortcutLabel: unref(translate)("search.keyboardShortcut"),
				onClick: handleClick
			}, {
				shortcut: withCtx(() => [isMac.value ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createElementVNode("span", _hoisted_3, toDisplayString(unref(translate)("search.command")), 1), _cache[0] || (_cache[0] = createElementVNode("span", { "aria-hidden": "true" }, "⌘", -1))], 64)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createElementVNode("span", _hoisted_4, toDisplayString(unref(translate)("search.control")), 1), _cache[1] || (_cache[1] = createElementVNode("span", { "aria-hidden": "true" }, "⌃", -1))], 64)), createTextVNode(" " + toDisplayString(__props.searchHotKey), 1)]),
				default: withCtx(() => [createElementVNode("span", _hoisted_1, toDisplayString(unref(translate)("search.open")), 1), createElementVNode("span", _hoisted_2, toDisplayString(unref(translate)("search.label")), 1)]),
				_: 1
			}, 8, ["class", "shortcutLabel"])), createVNode(SearchModal_default, {
				document: __props.document,
				eventBus: __props.eventBus,
				modalState: unref(modalState),
				modelsSectionLabel: __props.modelsSectionLabel
			}, null, 8, [
				"document",
				"eventBus",
				"modalState",
				"modelsSectionLabel"
			])], 64);
		};
	}
});
//#endregion
export { SearchButton_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SearchButton.vue.script.js.map