import type { AsyncApiChannelObject, AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
/**
 * Server names and protocols a channel is available on.
 *
 * Resolved from `document.servers`, restricted to `channel.servers` when the channel declares them.
 * `webSocketOnly` is disabled so labels cover every protocol, not just WebSocket, and protocols are
 * de-duplicated while preserving declaration order.
 */
export declare const getChannelServerLabels: (document: AsyncApiDocument, channel: AsyncApiChannelObject | null | undefined) => {
    servers: string[];
    protocols: string[];
};
//# sourceMappingURL=get-async-api-labels.d.ts.map