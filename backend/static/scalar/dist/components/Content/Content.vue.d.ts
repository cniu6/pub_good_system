import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { Heading } from '@scalar/types/legacy';
import type { AuthStore } from '@scalar/workspace-store/entities/auth';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { XScalarEnvironment } from '@scalar/workspace-store/schemas/extensions/document/x-scalar-environments';
import type { TraversedEntry as TraversedEntryType } from '@scalar/workspace-store/schemas/navigation';
import type { Workspace, WorkspaceDocument } from '@scalar/workspace-store/schemas/workspace';
type __VLS_Props = {
    infoSectionId: string;
    /** Slug of the active document, used to scope plugin view ids for navigation and deep-linking */
    documentSlug: string;
    /** The subset of the configuration object required for the content component */
    options: Pick<ApiReferenceConfigurationRaw, 'authentication' | 'baseServerURL' | 'documentDownloadType' | 'expandAllResponses' | 'hiddenClients' | 'hideTestRequestButton' | 'layout' | 'oauth2RedirectUri' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties' | 'persistAuth' | 'proxyUrl' | 'servers' | 'showOperationId' | 'hideModels' | 'modelsSectionLabel'>;
    document: WorkspaceDocument | undefined;
    clientDocument: WorkspaceDocument | undefined;
    authStore: AuthStore;
    xScalarDefaultClient: Workspace['x-scalar-default-client'];
    xScalarDefaultExample: Workspace['x-scalar-default-example'];
    items: TraversedEntryType[];
    expandedItems: Record<string, boolean>;
    eventBus: WorkspaceEventBus;
    environment: XScalarEnvironment;
    /** Heading id generator for Markdown headings */
    headingSlugGenerator: (heading: Heading) => string;
};
declare var __VLS_6: {}, __VLS_114: {};
type __VLS_Slots = {} & {
    start?: (props: typeof __VLS_6) => any;
} & {
    end?: (props: typeof __VLS_114) => any;
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
//# sourceMappingURL=Content.vue.d.ts.map