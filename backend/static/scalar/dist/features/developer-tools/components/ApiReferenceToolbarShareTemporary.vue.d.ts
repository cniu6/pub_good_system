import type { ExternalUrls } from '@scalar/types/api-reference';
import type { WorkspaceStore } from '@scalar/workspace-store/client';
type __VLS_Props = {
    workspace: WorkspaceStore;
    externalUrls: ExternalUrls;
};
type __VLS_ModelProps = {
    'url'?: string;
};
type __VLS_PublicProps = __VLS_Props & __VLS_ModelProps;
declare const __VLS_export: import("vue").DefineComponent<__VLS_PublicProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:url": (value: string | undefined) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_PublicProps> & Readonly<{
    "onUpdate:url"?: ((value: string | undefined) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ApiReferenceToolbarShareTemporary.vue.d.ts.map