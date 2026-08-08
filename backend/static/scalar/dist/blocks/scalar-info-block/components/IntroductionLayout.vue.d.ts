import type { AsyncApiInfoObject } from '@scalar/types/asyncapi/3.1';
import type { Heading } from '@scalar/types/legacy';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { ExternalDocumentationObject, InfoObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    id: string | undefined;
    documentType?: 'openapi' | 'asyncapi';
    specificationVersion: string | undefined;
    info: InfoObject | AsyncApiInfoObject | undefined;
    externalDocs?: ExternalDocumentationObject;
    documentExtensions?: Record<string, unknown>;
    infoExtensions?: Record<string, unknown>;
    headingSlugGenerator: (heading: Heading) => string;
    eventBus: WorkspaceEventBus | null;
};
declare var __VLS_67: {}, __VLS_80: {}, __VLS_92: {};
type __VLS_Slots = {} & {
    'download-link'?: (props: typeof __VLS_67) => any;
} & {
    aside?: (props: typeof __VLS_80) => any;
} & {
    after?: (props: typeof __VLS_92) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=IntroductionLayout.vue.d.ts.map