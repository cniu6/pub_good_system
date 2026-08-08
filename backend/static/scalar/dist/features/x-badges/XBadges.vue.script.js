import Badge_default from "../../components/Badge/Badge.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createTextVNode, defineComponent, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
//#region src/features/x-badges/XBadges.vue?vue&type=script&setup=true&lang.ts
var XBadges_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "XBadges",
	props: {
		position: {},
		badges: {}
	},
	setup(__props) {
		const filteredBadges = computed(() => {
			if (Array.isArray(__props.badges)) return __props.badges.filter((badge) => badge.position === __props.position || __props.position === "after" && !badge.position);
			return [];
		});
		return (_ctx, _cache) => {
			return filteredBadges.value.length ? (openBlock(true), createElementBlock(Fragment, { key: 0 }, renderList(filteredBadges.value, (badge) => {
				return openBlock(), createBlock(unref(Badge_default), {
					key: badge.name,
					color: badge.color
				}, {
					default: withCtx(() => [createTextVNode(toDisplayString(badge.name), 1)]),
					_: 2
				}, 1032, ["color"]);
			}), 128)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { XBadges_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=XBadges.vue.script.js.map