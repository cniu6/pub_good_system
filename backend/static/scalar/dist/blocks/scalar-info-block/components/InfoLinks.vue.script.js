import LinkList_default from "../../../components/LinkList/LinkList.vue.js";
import ExternalDocs_default from "../../../features/external-docs/ExternalDocs.vue.js";
import Contact_default from "../../../features/info-object/Contact.vue.js";
import InfoLink_default from "../../../features/info-object/InfoLink.vue.js";
import License_default from "../../../features/info-object/License.vue.js";
import TermsOfService_default from "../../../features/info-object/TermsOfService.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, renderList, unref, withCtx } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/blocks/scalar-info-block/components/InfoLinks.vue?vue&type=script&setup=true&lang.ts
var InfoLinks_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "InfoLinks",
	props: {
		info: {},
		externalDocs: {}
	},
	setup(__props) {
		/** Additional named links from the `x-scalar-links` extension (e.g. privacy policy, imprint) */
		const links = computed(() => {
			const value = __props.info["x-scalar-links"];
			if (!Array.isArray(value)) return [];
			return value.filter((link) => typeof link?.name === "string" && typeof link?.url === "string");
		});
		/** Whether there is at least one link to show, so we do not render an empty list */
		const hasLinks = computed(() => Boolean(__props.externalDocs || __props.info.contact || __props.info.license || __props.info.termsOfService || links.value.length));
		return (_ctx, _cache) => {
			return hasLinks.value ? (openBlock(), createBlock(unref(LinkList_default), { key: 0 }, {
				default: withCtx(() => [
					__props.externalDocs ? (openBlock(), createBlock(unref(ExternalDocs_default), {
						key: 0,
						value: __props.externalDocs
					}, null, 8, ["value"])) : createCommentVNode("", true),
					__props.info.contact ? (openBlock(), createBlock(unref(Contact_default), {
						key: 1,
						value: __props.info.contact
					}, null, 8, ["value"])) : createCommentVNode("", true),
					__props.info.license ? (openBlock(), createBlock(unref(License_default), {
						key: 2,
						value: unref(getResolvedRef)(__props.info.license)
					}, null, 8, ["value"])) : createCommentVNode("", true),
					__props.info.termsOfService ? (openBlock(), createBlock(unref(TermsOfService_default), {
						key: 3,
						value: __props.info.termsOfService
					}, null, 8, ["value"])) : createCommentVNode("", true),
					(openBlock(true), createElementBlock(Fragment, null, renderList(links.value, (link) => {
						return openBlock(), createBlock(unref(InfoLink_default), {
							key: link.url,
							name: link.name,
							url: link.url
						}, null, 8, ["name", "url"]);
					}), 128))
				]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { InfoLinks_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=InfoLinks.vue.script.js.map