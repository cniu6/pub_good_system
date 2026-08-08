import { useLocalization } from "../../../features/localization/use-localization.js";
import Badge_default from "../../Badge/Badge.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, defineComponent, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
//#region src/components/Content/AsyncApi/AsyncApiLabels.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "async-api-labels"
};
var _hoisted_2 = { class: "sr-only" };
var _hoisted_3 = { class: "sr-only" };
var AsyncApiLabels_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AsyncApiLabels",
	props: {
		servers: { default: () => [] },
		protocols: { default: () => [] }
	},
	setup(__props) {
		const { translate } = useLocalization();
		/** Hide the whole row when there is nothing to show. */
		const hasLabels = computed(() => __props.servers.length > 0 || __props.protocols.length > 0);
		return (_ctx, _cache) => {
			return hasLabels.value ? (openBlock(), createElementBlock("div", _hoisted_1, [__props.servers.length ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createElementVNode("span", _hoisted_2, toDisplayString(unref(translate)("asyncapi.servers")) + ":", 1), (openBlock(true), createElementBlock(Fragment, null, renderList(__props.servers, (server) => {
				return openBlock(), createBlock(unref(Badge_default), {
					key: `server-${server}`,
					class: "async-api-label--server",
					title: server
				}, {
					default: withCtx(() => [createTextVNode(toDisplayString(server), 1)]),
					_: 2
				}, 1032, ["title"]);
			}), 128))], 64)) : createCommentVNode("", true), __props.protocols.length ? (openBlock(), createElementBlock(Fragment, { key: 1 }, [createElementVNode("span", _hoisted_3, toDisplayString(unref(translate)("asyncapi.protocols")) + ":", 1), (openBlock(true), createElementBlock(Fragment, null, renderList(__props.protocols, (protocol) => {
				return openBlock(), createBlock(unref(Badge_default), {
					key: `protocol-${protocol}`,
					class: "async-api-label--protocol",
					title: protocol
				}, {
					default: withCtx(() => [createTextVNode(toDisplayString(protocol), 1)]),
					_: 2
				}, 1032, ["title"]);
			}), 128))], 64)) : createCommentVNode("", true)])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { AsyncApiLabels_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AsyncApiLabels.vue.script.js.map