import { useLocalization } from "../../localization/use-localization.js";
import SecurityRequirementBadgeScheme_default from "./SecurityRequirementBadgeScheme.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, onBeforeUnmount, openBlock, ref, renderList, toDisplayString, unref, withCtx, withModifiers } from "vue";
import { onClickOutside, onKeyStroke } from "@vueuse/core";
import { ScalarIconLockSimple, ScalarIconLockSimpleOpen } from "@scalar/icons";
import { ScalarFloating, ScalarFloatingBackdrop } from "@scalar/components/floating";
//#region src/features/Operation/components/SecurityRequirementBadge.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-expanded"];
var _hoisted_2 = { key: 2 };
var _hoisted_3 = { class: "flex max-w-xs min-w-48 flex-col gap-1.5 p-2 text-sm" };
var _hoisted_4 = { class: "font-medium" };
var _hoisted_5 = {
	key: 0,
	class: "contents"
};
var _hoisted_6 = {
	key: 1,
	class: "contents"
};
var _hoisted_7 = {
	key: 1,
	class: "contents"
};
var SecurityRequirementBadge_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SecurityRequirementBadge",
	props: {
		requiredSecurity: {},
		hideLabel: {
			type: Boolean,
			default: false
		}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* The badge shows a small panel with the security details. It opens on hover
		* and on click, so we own the open state directly instead of leaning on a
		* click-only popover. Owning the state keeps the two interactions from fighting
		* each other: hover and the click of the same gesture (a tap fires both) can no
		* longer toggle each other off.
		*/
		const triggerRef = ref(null);
		const panelRef = ref(null);
		const isOpen = ref(false);
		/**
		* Whether the panel is pinned open by a click. A pinned panel ignores the
		* pointer leaving so it behaves like the old click-to-open popover, and it only
		* closes on another click, a click outside, or Escape.
		*/
		const isPinned = ref(false);
		let closeTimeout;
		/**
		* Close after a short delay so the pointer can travel across the gap between the
		* badge and the panel without the panel disappearing. A pinned panel stays open.
		*/
		const scheduleClose = () => {
			if (isPinned.value) return;
			clearTimeout(closeTimeout);
			closeTimeout = setTimeout(() => {
				isOpen.value = false;
			}, 120);
		};
		const cancelClose = () => clearTimeout(closeTimeout);
		const openOnHover = () => {
			cancelClose();
			isOpen.value = true;
		};
		const close = () => {
			cancelClose();
			isOpen.value = false;
			isPinned.value = false;
		};
		/**
		* Toggle on click. A hover already opened the panel (and a tap's `mouseenter`
		* fires just before its `click`), so the first click pins it open rather than
		* closing it; a click on an already pinned panel closes it.
		*/
		const toggleOnClick = () => {
			if (isPinned.value) {
				close();
				return;
			}
			cancelClose();
			isOpen.value = true;
			isPinned.value = true;
		};
		onClickOutside(triggerRef, close, { ignore: [panelRef] });
		onKeyStroke("Escape", () => {
			if (isOpen.value) close();
		});
		onBeforeUnmount(() => clearTimeout(closeTimeout));
		const label = computed(() => __props.requiredSecurity.state === "required" ? translate("authentication.required") : translate("authentication.optional"));
		const verb = computed(() => __props.requiredSecurity.state === "required" ? translate("authentication.requires") : translate("authentication.accepts"));
		/** Single group, single scheme — shown inline in the header. */
		const isSingleScheme = computed(() => __props.requiredSecurity.requirements.length === 1 && __props.requiredSecurity.requirements[0]?.schemes.length === 1);
		/** Single group with multiple schemes — all must be satisfied (AND). */
		const isAndGroup = computed(() => __props.requiredSecurity.requirements.length === 1 && (__props.requiredSecurity.requirements[0]?.schemes.length ?? 0) > 1);
		/** Multiple groups — any one group satisfies authentication (OR). */
		const isOrAlternatives = computed(() => __props.requiredSecurity.requirements.length > 1);
		return (_ctx, _cache) => {
			return __props.requiredSecurity.state !== "none" ? (openBlock(), createBlock(unref(ScalarFloating), {
				key: 0,
				placement: "bottom-end"
			}, {
				floating: withCtx(() => [isOpen.value ? (openBlock(), createElementBlock("div", {
					key: 0,
					ref_key: "panelRef",
					ref: panelRef,
					class: "relative flex flex-col p-0.75",
					onClick: _cache[0] || (_cache[0] = withModifiers(() => {}, ["stop"])),
					onMouseenter: cancelClose,
					onMouseleave: scheduleClose
				}, [createElementVNode("div", _hoisted_3, [createElementVNode("div", _hoisted_4, [createTextVNode(toDisplayString(verb.value) + " ", 1), isSingleScheme.value ? (openBlock(), createBlock(SecurityRequirementBadgeScheme_default, {
					key: 0,
					is: "span",
					class: "contents",
					scheme: __props.requiredSecurity.requirements[0].schemes[0]
				}, null, 8, ["scheme"])) : isOrAlternatives.value ? (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(unref(translate)("authentication.oneOf")), 1)], 64)) : isAndGroup.value ? (openBlock(), createElementBlock(Fragment, { key: 2 }, [createTextVNode(toDisplayString(unref(translate)("authentication.allOf")), 1)], 64)) : (openBlock(), createElementBlock(Fragment, { key: 3 }, [createTextVNode(toDisplayString(unref(translate)("authentication.authentication")), 1)], 64))]), isOrAlternatives.value ? (openBlock(), createElementBlock("ul", _hoisted_5, [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.requiredSecurity.requirements, (group, gi) => {
					return openBlock(), createElementBlock("li", {
						key: gi,
						class: "markdown"
					}, [group.schemes.length === 1 ? (openBlock(), createBlock(SecurityRequirementBadgeScheme_default, {
						key: 0,
						is: "span",
						class: "contents",
						scheme: group.schemes[0]
					}, null, 8, ["scheme"])) : (openBlock(), createElementBlock("ul", _hoisted_6, [(openBlock(true), createElementBlock(Fragment, null, renderList(group.schemes, (scheme, si) => {
						return openBlock(), createBlock(SecurityRequirementBadgeScheme_default, {
							key: si,
							scheme
						}, null, 8, ["scheme"]);
					}), 128))]))]);
				}), 128))])) : isAndGroup.value ? (openBlock(), createElementBlock("ul", _hoisted_7, [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.requiredSecurity.requirements[0].schemes, (scheme, key) => {
					return openBlock(), createBlock(SecurityRequirementBadgeScheme_default, {
						key,
						scheme
					}, null, 8, ["scheme"]);
				}), 128))])) : createCommentVNode("", true)]), createVNode(unref(ScalarFloatingBackdrop))], 544)) : createCommentVNode("", true)]),
				default: withCtx(() => [createElementVNode("button", {
					ref_key: "triggerRef",
					ref: triggerRef,
					class: normalizeClass(["security-requirement-badge inline-flex w-fit shrink-0 items-center justify-center gap-1 text-sm", __props.requiredSecurity.state === "optional" ? "text-c-2" : "text-c-1 font-medium"]),
					type: "button",
					"aria-expanded": isOpen.value,
					"aria-haspopup": "dialog",
					onClick: withModifiers(toggleOnClick, ["stop"]),
					onMouseenter: openOnHover,
					onMouseleave: scheduleClose
				}, [__props.requiredSecurity.state === "required" ? (openBlock(), createBlock(unref(ScalarIconLockSimple), {
					key: 0,
					class: "size-3",
					weight: "bold"
				})) : (openBlock(), createBlock(unref(ScalarIconLockSimpleOpen), {
					key: 1,
					class: "size-3",
					weight: "bold"
				})), !__props.hideLabel ? (openBlock(), createElementBlock("span", _hoisted_2, toDisplayString(label.value), 1)) : createCommentVNode("", true)], 42, _hoisted_1)]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SecurityRequirementBadge_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SecurityRequirementBadge.vue.script.js.map