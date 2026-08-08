import Badge_default from "../../../components/Badge/Badge.vue.js";
import { computed, createBlock, createCommentVNode, createTextVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
//#region src/blocks/scalar-info-block/components/InfoVersion.vue?vue&type=script&setup=true&lang.ts
var InfoVersion_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoVersion",
	props: { version: {} },
	setup(__props) {
		/** Format the version number to be displayed in the badge */
		const prefixedVersion = computed(() => {
			if (__props.version == null) return __props.version;
			const versionString = String(__props.version);
			return /^\d/.test(versionString) ? `v${versionString}` : versionString;
		});
		return (_ctx, _cache) => {
			return prefixedVersion.value ? (openBlock(), createBlock(unref(Badge_default), { key: 0 }, {
				default: withCtx(() => [createTextVNode(toDisplayString(prefixedVersion.value), 1)]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { InfoVersion_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoVersion.vue.script.js.map