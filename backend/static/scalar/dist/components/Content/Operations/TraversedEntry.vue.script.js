import SectionContainer_default from "../../Section/SectionContainer.vue.js";
import Model_default from "../Models/Model.vue.js";
import ModelTag_default from "../Models/ModelTag.vue.js";
import Tag_default from "../Tags/Tag.vue.js";
import Lazy_default from "../../Lazy/Lazy.vue.js";
import Operation_default from "../../../features/Operation/Operation.vue.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createVNode, defineComponent, openBlock, renderList, resolveComponent, unref, withCtx } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { getResolvedPathItem } from "@scalar/workspace-store/helpers/for-each-path-item-operation";
//#region src/components/Content/Operations/TraversedEntry.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var TraversedEntry_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "TraversedEntry",
	props: {
		authStore: {},
		level: { default: 0 },
		entries: {},
		document: {},
		clientOptions: {},
		options: {},
		selectedServer: {},
		securitySchemes: {},
		selectedClient: {},
		selectedExample: {},
		expandedItems: {},
		eventBus: {}
	},
	setup(__props) {
		/**
		* Type guards for different entry types
		*/
		const isTagGroup = (entry) => entry["type"] === "tag" && entry.isGroup === true;
		const isTag = (entry) => entry["type"] === "tag" && !isTagGroup(entry) && entry.id !== "models";
		const isOperation = (entry) => entry["type"] === "operation";
		const isWebhook = (entry) => entry["type"] === "webhook";
		/** Models are special form of tag entry */
		const isModelsTag = (entry) => entry["type"] === "models";
		const isModel = (entry) => entry["type"] === "model";
		function getPathValue(entry) {
			return isWebhook(entry) ? getResolvedPathItem(__props.document.webhooks?.[entry.name]) : getResolvedPathItem(__props.document.paths?.[entry.path]);
		}
		return (_ctx, _cache) => {
			const _component_TraversedEntry = resolveComponent("TraversedEntry", true);
			return openBlock(true), createElementBlock(Fragment, null, renderList(__props.entries, (entry) => {
				return openBlock(), createBlock(Lazy_default, {
					id: entry.id,
					key: `${entry.id}-${__props.options.layout}`,
					expanded: !!__props.expandedItems[entry.id]
				}, {
					default: withCtx(() => [isOperation(entry) || isWebhook(entry) ? (openBlock(), createBlock(unref(SectionContainer_default), {
						key: 0,
						omit: __props.level !== 0
					}, {
						default: withCtx(() => [createVNode(unref(Operation_default), {
							id: entry.id,
							authStore: __props.authStore,
							clientOptions: __props.clientOptions,
							document: __props.document,
							eventBus: __props.eventBus,
							isCollapsed: !__props.expandedItems[entry.id],
							isWebhook: isWebhook(entry),
							method: entry.method,
							options: __props.options,
							path: isWebhook(entry) ? entry.name : entry.path,
							pathValue: getPathValue(entry),
							securitySchemes: __props.securitySchemes,
							selectedClient: __props.selectedClient,
							selectedExample: __props.selectedExample,
							server: __props.selectedServer
						}, null, 8, [
							"id",
							"authStore",
							"clientOptions",
							"document",
							"eventBus",
							"isCollapsed",
							"isWebhook",
							"method",
							"options",
							"path",
							"pathValue",
							"securitySchemes",
							"selectedClient",
							"selectedExample",
							"server"
						])]),
						_: 2
					}, 1032, ["omit"])) : isTag(entry) || isTagGroup(entry) && __props.options.layout === "classic" ? (openBlock(), createBlock(unref(Tag_default), {
						key: 1,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						layout: __props.options.layout,
						moreThanOneTag: __props.entries.filter(isTag).length > 1,
						tag: entry
					}, {
						default: withCtx(() => ["children" in entry && entry.children?.length ? (openBlock(), createBlock(_component_TraversedEntry, {
							key: 0,
							authStore: __props.authStore,
							clientOptions: __props.clientOptions,
							document: __props.document,
							entries: entry.children,
							eventBus: __props.eventBus,
							expandedItems: __props.expandedItems,
							level: __props.level + 1,
							options: __props.options,
							securitySchemes: __props.securitySchemes,
							selectedClient: __props.selectedClient,
							selectedExample: __props.selectedExample,
							selectedServer: __props.selectedServer
						}, null, 8, [
							"authStore",
							"clientOptions",
							"document",
							"entries",
							"eventBus",
							"expandedItems",
							"level",
							"options",
							"securitySchemes",
							"selectedClient",
							"selectedExample",
							"selectedServer"
						])) : createCommentVNode("", true)]),
						_: 2
					}, 1032, [
						"eventBus",
						"isCollapsed",
						"layout",
						"moreThanOneTag",
						"tag"
					])) : isTagGroup(entry) ? (openBlock(), createElementBlock("div", {
						key: 2,
						id: entry.id
					}, [createVNode(_component_TraversedEntry, {
						authStore: __props.authStore,
						clientOptions: __props.clientOptions,
						document: __props.document,
						entries: entry.children || [],
						eventBus: __props.eventBus,
						expandedItems: __props.expandedItems,
						level: __props.level + 1,
						options: __props.options,
						securitySchemes: __props.securitySchemes,
						selectedClient: __props.selectedClient,
						selectedExample: __props.selectedExample,
						selectedServer: __props.selectedServer
					}, null, 8, [
						"authStore",
						"clientOptions",
						"document",
						"entries",
						"eventBus",
						"expandedItems",
						"level",
						"options",
						"securitySchemes",
						"selectedClient",
						"selectedExample",
						"selectedServer"
					])], 8, _hoisted_1)) : isModelsTag(entry) && __props.document.components?.schemas ? (openBlock(), createBlock(ModelTag_default, {
						key: 3,
						id: entry.id,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						layout: __props.options.layout,
						modelsSectionLabel: __props.options.modelsSectionLabel
					}, {
						default: withCtx(() => [createVNode(_component_TraversedEntry, {
							authStore: __props.authStore,
							clientOptions: __props.clientOptions,
							document: __props.document,
							entries: entry.children || [],
							eventBus: __props.eventBus,
							expandedItems: __props.expandedItems,
							level: __props.level + 1,
							options: __props.options,
							securitySchemes: __props.securitySchemes,
							selectedClient: __props.selectedClient,
							selectedExample: __props.selectedExample,
							selectedServer: __props.selectedServer
						}, null, 8, [
							"authStore",
							"clientOptions",
							"document",
							"entries",
							"eventBus",
							"expandedItems",
							"level",
							"options",
							"securitySchemes",
							"selectedClient",
							"selectedExample",
							"selectedServer"
						])]),
						_: 2
					}, 1032, [
						"id",
						"eventBus",
						"isCollapsed",
						"layout",
						"modelsSectionLabel"
					])) : isModel(entry) && __props.document.components?.schemas?.[entry.name] ? (openBlock(), createBlock(Model_default, {
						key: 4,
						id: entry.id,
						document: __props.document,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						name: entry.name,
						options: __props.options,
						schema: unref(getResolvedRef)(__props.document.components.schemas[entry.name])
					}, null, 8, [
						"id",
						"document",
						"eventBus",
						"isCollapsed",
						"name",
						"options",
						"schema"
					])) : createCommentVNode("", true)]),
					_: 2
				}, 1032, ["id", "expanded"]);
			}), 128);
		};
	}
});
//#endregion
export { TraversedEntry_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=TraversedEntry.vue.script.js.map