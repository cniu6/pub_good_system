<template>
  <n-card :title="t('adminSettings.title')" :bordered="false">
    <n-spin :show="loading">
      <n-tabs v-model:value="topTab" type="line" animated>
        <n-tab-pane name="system-config" :tab="t('adminSettings.systemConfig')">
          <n-tabs v-model:value="systemSubTab" type="line" placement="left" animated>
            <n-tab-pane name="basic" :tab="t('adminSettings.basicSettings')">
              <n-space vertical>
                <n-form :model="basicForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.siteName')">
                    <n-input v-model:value="basicForm.site_name" :placeholder="t('adminSettings.siteNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.siteDesc')">
                    <n-input v-model:value="basicForm.site_desc" type="textarea" :placeholder="t('adminSettings.siteDescPlaceholder')" :rows="3" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.siteLogo')">
                    <n-input v-model:value="basicForm.site_logo" :placeholder="t('adminSettings.siteLogoPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.copyright')">
                    <n-input v-model:value="basicForm.copyright" :placeholder="t('adminSettings.copyrightPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.icp')">
                    <n-input v-model:value="basicForm.icp" :placeholder="t('adminSettings.icpPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.version')">
                    <n-input v-model:value="basicForm.version" :placeholder="t('adminSettings.versionPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.defaultLang')">
                    <n-select v-model:value="basicForm.default_lang" :options="langOptions" :placeholder="t('adminSettings.defaultLangPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.frontendUrl')">
                    <n-input v-model:value="basicForm.frontend_url" :placeholder="t('adminSettings.frontendUrlPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.backendApiUrl')">
                    <n-input v-model:value="basicForm.backend_api_url" :placeholder="t('adminSettings.backendApiUrlPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.userRegistration')">
                    <n-space align="center">
                      <n-switch
                        :value="basicForm.allow_register"
                        :loading="switchLoading.allow_register"
                        @update:value="handleUpdateAllowRegister"
                      />
                      <n-text depth="3">{{ basicForm.allow_register ? t('adminSettings.allowRegister') : t('adminSettings.disallowRegister') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingBasic" @click="handleSaveBasic">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="email" :tab="t('adminSettings.emailSettings')">
              <n-space vertical>
                <n-form :model="emailForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.emailVerification')">
                    <n-space align="center">
                      <n-switch
                        :value="emailForm.email_verify_enabled"
                        :loading="switchLoading.email_verify_enabled"
                        @update:value="handleUpdateEmailVerifyEnabled"
                      />
                      <n-text depth="3">{{ emailForm.email_verify_enabled ? t('adminSettings.emailVerifyEnabled') : t('adminSettings.emailVerifyDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.smtpHost')">
                    <n-input v-model:value="emailForm.smtp_host" :placeholder="t('adminSettings.smtpHostPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpPort')">
                    <n-input-number v-model:value="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpUsername')">
                    <n-input v-model:value="emailForm.smtp_username" :placeholder="t('adminSettings.smtpUsernamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpPassword')">
                    <n-input
                      v-model:value="emailForm.smtp_password"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.smtpPasswordPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.systemEmailName')">
                    <n-input v-model:value="emailForm.system_email_name" :placeholder="t('adminSettings.systemEmailNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smtpSSL')">
                    <n-space align="center">
                      <n-switch :value="emailForm.smtp_ssl" :loading="switchLoading.smtp_ssl" @update:value="handleUpdateSmtpSSL" />
                      <n-text depth="3">{{ emailForm.smtp_ssl ? t('adminSettings.sslEnabled') : t('adminSettings.sslDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingEmail" @click="handleSaveEmail">{{ t('adminSettings.save') }}</n-button>
                      <n-button :loading="testingEmail" @click="handleTestEmail">{{ t('adminSettings.sendTestEmail') }}</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="sms" :tab="t('adminSettings.smsSettings')">
              <n-space vertical>
                <n-form :model="smsForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.smsVerification')">
                    <n-space align="center">
                      <n-switch
                        :value="smsForm.sms_verify_enabled"
                        :loading="switchLoading.sms_verify_enabled"
                        @update:value="handleUpdateSmsVerifyEnabled"
                      />
                      <n-text depth="3">{{ smsForm.sms_verify_enabled ? t('adminSettings.smsVerifyEnabled') : t('adminSettings.smsVerifyDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.smsProvider')">
                    <n-select
                      v-model:value="smsForm.sms_provider"
                      :options="smsProviderOptions"
                      :placeholder="t('adminSettings.smsProviderPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smsAccessKey')">
                    <n-input
                      v-model:value="smsForm.sms_access_key"
                      :placeholder="smsAccessKeyPlaceholder"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.smsSecretKey')">
                    <n-input
                      v-model:value="smsForm.sms_secret_key"
                      type="password"
                      show-password-on="click"
                      :placeholder="smsSecretKeyPlaceholder"
                    />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsSignName" :label="t('adminSettings.smsSignName')">
                    <n-input v-model:value="smsForm.sms_sign_name" :placeholder="t('adminSettings.smsSignNamePlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateLabel">
                    <n-input v-model:value="smsForm.sms_template_code" :placeholder="smsTemplatePlaceholder" />
                  </n-form-item>
                  <n-form-item v-if="smsProviderNeedsTemplateCode" :label="smsTemplateEnLabel">
                    <n-input v-model:value="smsForm.sms_template_code_en" :placeholder="smsTemplateEnPlaceholder" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'aliyun'" :label="t('adminSettings.smsRegion')">
                    <n-input v-model:value="smsForm.sms_region" :placeholder="t('adminSettings.smsRegionPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'tencent'" :label="t('adminSettings.smsSdkAppId')">
                    <n-input v-model:value="smsForm.sms_sdk_app_id" :placeholder="t('adminSettings.smsSdkAppIdPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsEndpoint')">
                    <n-input v-model:value="smsForm.sms_endpoint" :placeholder="t('adminSettings.smsEndpointPlaceholder')" />
                  </n-form-item>
                  <n-form-item v-if="smsForm.sms_provider === 'custom'" :label="t('adminSettings.smsBodyFormat')">
                    <n-select
                      v-model:value="smsForm.sms_body_format"
                      :options="smsBodyFormatOptions"
                      :placeholder="t('adminSettings.smsBodyFormatPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingSms" @click="handleSaveSms">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
                  {{ t('adminSettings.smsAlert') }}
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="security" :tab="t('adminSettings.securitySettings')">
              <n-space vertical>
                <n-form :model="securityForm" label-placement="left" label-width="180px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.geetestEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.geetest_enabled"
                        :loading="switchLoading.geetest_enabled"
                        @update:value="handleUpdateGeetestEnabled"
                      />
                      <n-text depth="3">{{ securityForm.geetest_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.allowDeleteAccount')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.allow_delete_account"
                        :loading="switchLoading.allow_delete_account"
                        @update:value="handleUpdateAllowDeleteAccount"
                      />
                      <n-text depth="3">{{ securityForm.allow_delete_account ? t('adminSettings.allowDeleteAccountEnabled') : t('adminSettings.allowDeleteAccountDisabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.geetestCaptchaId')">
                    <n-input v-model:value="securityForm.geetest_captcha_id" :placeholder="t('adminSettings.geetestCaptchaIdPlaceholder')" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.geetestCaptchaKey')">
                    <n-input
                      v-model:value="securityForm.geetest_captcha_key"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.geetestCaptchaKeyPlaceholder')"
                    />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.jwtAccessExpire')">
                    <n-input-number v-model:value="securityForm.jwt_access_expire" :min="300" :step="300" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.jwtRefreshExpire')">
                    <n-input-number v-model:value="securityForm.jwt_refresh_expire" :min="3600" :step="3600" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.loginMaxFailure')">
                    <n-input-number v-model:value="securityForm.login_max_failure" :min="3" :max="20" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.loginLockDuration')">
                    <n-input-number v-model:value="securityForm.login_lock_duration" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.realnameEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.realname_enabled"
                        :loading="switchLoading.realname_enabled"
                        @update:value="handleUpdateRealnameEnabled"
                      />
                      <n-text depth="3">{{ securityForm.realname_enabled ? t('adminSettings.realnameEnabledText') : t('adminSettings.realnameDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameReview')">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.realname_review_required"
                        :loading="switchLoading.realname_review_required"
                        @update:value="handleUpdateRealnameReviewRequired"
                      />
                      <n-text depth="3">{{ securityForm.realname_review_required ? t('adminSettings.realnameReviewRequired') : t('adminSettings.realnameReviewNotRequired') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameNotifyText')">
                    <n-input
                      v-model:value="securityForm.realname_notify_text"
                      type="textarea"
                      :placeholder="t('adminSettings.realnameNotifyTextPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingSecurity" @click="handleSaveSecurity">{{ t('adminSettings.saveSettings') }}</n-button>
                      <n-button type="warning" :loading="restartingBackend" @click="handleRestartBackend">{{ t('adminSettings.restartBackend') }}</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="realname-api" :tab="t('adminSettings.realnameApiSettings')">
              <n-space vertical>
                <n-form :model="realnameApiForm" label-placement="left" label-width="180px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.realnameApiEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="realnameApiForm.realname_api_enabled"
                        :loading="switchLoading.realname_api_enabled"
                        @update:value="handleUpdateRealnameApiEnabled"
                      />
                      <n-text depth="3">{{ realnameApiForm.realname_api_enabled ? t('adminSettings.realnameApiEnabledText') : t('adminSettings.realnameApiDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.realnameApiProvider')">
                    <n-select
                      v-model:value="realnameApiForm.realname_api_provider"
                      :options="realnameApiProviderOptions"
                      :placeholder="t('adminSettings.realnameApiProviderPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiAppKey')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_app_key"
                      :placeholder="t('adminSettings.realnameApiAppKeyPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiAppSecret')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_app_secret"
                      type="password"
                      show-password-on="click"
                      :placeholder="t('adminSettings.realnameApiAppSecretPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.realnameApiEndpoint')">
                    <n-input
                      v-model:value="realnameApiForm.realname_api_endpoint"
                      :placeholder="t('adminSettings.realnameApiEndpointPlaceholder')"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingRealnameApi" @click="handleSaveRealnameApi">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.tip')" :bordered="false">
                  <p>{{ t('adminSettings.realnameApiAlert1') }}</p>
                  <p>{{ t('adminSettings.realnameApiAlert2') }}</p>
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="payment" :tab="t('adminSettings.paymentSettings')">
              <n-space vertical>
                <n-form :model="paymentForm" label-placement="left" label-width="140px" style="max-width: 640px;">
                  <n-form-item :label="t('adminSettings.paymentEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="paymentForm.payment_enabled"
                        :loading="switchLoading.payment_enabled"
                        @update:value="handleUpdatePaymentEnabled"
                      />
                      <n-text depth="3">{{ paymentForm.payment_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.orderExpireMinutes')">
                    <n-input-number v-model:value="paymentForm.payment_order_expire_minutes" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-divider />
                  <n-form-item :label="t('adminSettings.withdrawEnabled')">
                    <n-space align="center">
                      <n-switch
                        :value="paymentForm.withdraw_enabled"
                        :loading="switchLoading.withdraw_enabled"
                        @update:value="handleUpdateWithdrawEnabled"
                      />
                      <n-text depth="3">{{ paymentForm.withdraw_enabled ? t('adminSettings.withdrawEnabledText') : t('adminSettings.withdrawDisabledText') }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawMinAmount')">
                    <n-input-number v-model:value="paymentForm.withdraw_min_amount" :min="0.01" :precision="2" :step="1" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawNotifyText')">
                    <n-input
                      v-model:value="paymentForm.withdraw_notify_text"
                      type="textarea"
                      :placeholder="t('adminSettings.withdrawNotifyTextPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item :label="t('adminSettings.withdrawAccountTypes')">
                    <n-input
                      v-model:value="paymentForm.withdraw_account_types_text"
                      type="textarea"
                      :placeholder="t('adminSettings.withdrawAccountTypesPlaceholder')"
                      :rows="3"
                    />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingPayment" @click="handleSavePayment">{{ t('adminSettings.saveSettings') }}</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" :title="t('adminSettings.configDesc')" :bordered="false">
                  <ul style="margin: 0; padding-left: 18px;">
                    <li>{{ t('adminSettings.paymentAlert1') }}</li>
                    <li>{{ t('adminSettings.paymentAlert2') }}</li>
                  </ul>
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="custom" :tab="t('adminSettings.customConfig')">
              <n-space vertical :size="16">
                <n-space justify="end">
                  <TableColumnSelector
                    v-model="customSelectedColumnKeys"
                    :options="customColumnOptions"
                    :visible-count="customVisibleColumnCount"
                    :total-count="customTotalColumnCount"
                    :button-label="t('common.showFields')"
                    :title="t('common.visibleFields')"
                    :hint="t('common.columnVisibilityHint')"
                    :reset-label="t('common.restoreDefaultFields')"
                    @reset="resetCustomSelectedColumns"
                  />
                  <n-button type="primary" @click="showAddModal = true">
                    <template #icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" style="width: 1em; height: 1em;">
                        <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" />
                      </svg>
                    </template>
                    {{ t('adminSettings.addConfigItem') }}
                  </n-button>
                </n-space>

                <n-data-table :columns="customVisibleColumns" :data="customSettings" :pagination="false" :bordered="false" :scroll-x="customTableScrollX" />
              </n-space>
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>

        <n-tab-pane name="email-templates" :tab="t('adminSettings.emailTemplates')">
          <EmailTemplates />
        </n-tab-pane>

        <n-tab-pane v-if="false" name="sms-logs" :tab="t('adminSettings.smsLogs')">
          <SMSLogs />
        </n-tab-pane>

        <n-tab-pane v-if="false" name="operation-logs" :tab="t('adminSettings.operationLogs')">
          <OperationLogs />
        </n-tab-pane>

        <n-tab-pane v-if="false" name="info" :tab="t('adminSettings.systemInfo')">
          <n-space vertical>
            <n-descriptions bordered :column="2" label-placement="left">
              <n-descriptions-item :label="t('adminSettings.systemVersion')">{{ settingsStore.version }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.goVersion')">{{ t('adminSettings.goVersionValue') }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.frontendFramework')">{{ t('adminSettings.frontendFrameworkValue') }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.uiLibrary')">{{ t('adminSettings.uiLibraryValue') }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.database')">{{ t('adminSettings.databaseValue') }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.runtimeEnv')">{{ mode }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.nodeVersion')">{{ t('adminSettings.nodeVersionValue') }}</n-descriptions-item>
              <n-descriptions-item :label="t('adminSettings.buildTime')">{{ buildTime }}</n-descriptions-item>
            </n-descriptions>

            <n-card :title="t('adminSettings.loadedPlugins')" size="small">
              <n-space vertical>
                <n-tag v-for="p in plugins" :key="p.name" :type="p.active ? 'success' : 'default'" size="medium">
                  {{ p.name }} ({{ p.version }})
                </n-tag>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>

        <n-tab-pane v-if="false" name="server-management" :tab="t('adminSettings.serverManagement')">
          <n-tabs type="line" animated>
            <n-tab-pane name="monitor" :tab="t('adminSettings.systemMonitor')">
              <n-space vertical :size="16">
                <n-space justify="space-between" align="center">
                  <n-text depth="3">{{ t('adminSettings.realtimeMonitor') }}</n-text>
                  <n-button :loading="loadingServerMonitoring" @click="loadServerMonitoringStatus">{{ t('adminSettings.refresh') }}</n-button>
                </n-space>

                <n-grid :x-gap="10" :y-gap="10" cols="2 s:4 m:4 l:8" responsive="screen">
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.cpu')">
                        <template #default>
                          <n-text :type="cpuPercent > 80 ? 'error' : 'success'">{{ formatPercent(cpuPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ t('adminSettings.cpuCores', { count: serverMonitoringData?.metrics.cpu.core_count || 0 }) }}</n-text></template>
                      </n-statistic>
                      <n-progress type="line" :percentage="cpuPercent" :status="cpuPercent > 80 ? 'error' : 'success'" :show-indicator="false" style="margin-top: 8px" />
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.systemMemory')">
                        <template #default>
                          <n-text :type="memoryPercent > 80 ? 'error' : 'success'">{{ formatPercent(memoryPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromMB(serverMonitoringData?.metrics.memory.used_mb || 0) }}/{{ formatStorageFromMB(serverMonitoringData?.metrics.memory.total_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.swap')">
                        <template #default>
                          <n-text :type="swapPercent > 80 ? 'error' : 'success'">{{ formatPercent(swapPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromMB(serverMonitoringData?.metrics.swap.used_mb || 0) }}/{{ formatStorageFromMB(serverMonitoringData?.metrics.swap.total_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.processMemory')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.process_rss_mb || 0) }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">CPU {{ Number((serverMonitoringData?.process.process_cpu || 0).toFixed(2)) }}%</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.goHeap')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_alloc_mb || 0) }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">sys {{ formatStorageFromMB(serverMonitoringData?.process.memory_sys_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.goroutineGC')">
                        <template #default>{{ serverMonitoringData?.process.goroutines || 0 }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">GC {{ serverMonitoringData?.process.gc_count || 0 }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.diskUsage')">
                        <template #default>
                          <n-text :type="diskPercent > 80 ? 'error' : 'success'">{{ formatPercent(diskPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromGB(serverMonitoringData?.metrics.disk.used_gb || 0) }}/{{ formatStorageFromGB(serverMonitoringData?.metrics.disk.total_gb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic :label="t('adminSettings.uptime')">
                        <template #default>{{ uptimeTextPrecise }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ t('adminSettings.started') }}: {{ startTimeText }} · {{ uptimeText }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-space vertical size="small">
                        <n-statistic :label="t('adminSettings.network')">
                          <template #default>{{ formatBytes((serverMonitoringData?.metrics.network.bytes_sent || 0) + (serverMonitoringData?.metrics.network.bytes_recv || 0)) }}</template>
                        </n-statistic>
                        <n-space justify="space-between"><n-text depth="3">{{ t('adminSettings.upload') }}</n-text><n-text>{{ formatBytes(serverMonitoringData?.metrics.network.bytes_sent || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">{{ t('adminSettings.download') }}</n-text><n-text>{{ formatBytes(serverMonitoringData?.metrics.network.bytes_recv || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">{{ t('adminSettings.uploadPackets') }}</n-text><n-text>{{ formatInteger(serverMonitoringData?.metrics.network.packets_sent || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">{{ t('adminSettings.downloadPackets') }}</n-text><n-text>{{ formatInteger(serverMonitoringData?.metrics.network.packets_recv || 0) }}</n-text></n-space>
                      </n-space>
                    </n-card>
                  </n-gi>
                </n-grid>

                <n-card size="small" :title="t('adminSettings.memoryDetails')">
                  <n-grid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
                    <n-gi>
                      <n-statistic :label="t('adminSettings.goMemoryAlloc')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.memory_alloc_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.goMemorySys')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.memory_sys_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.heapAlloc')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_alloc_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.heapInUse')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_inuse_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.heapIdle')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_idle_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.stackInUse')">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.stack_inuse_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.gcCount')">
                        <template #default>{{ serverMonitoringData?.process.gc_count || 0 }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic :label="t('adminSettings.gcCPU')">
                        <template #default>{{ Number(((serverMonitoringData?.process.gc_cpu_fraction || 0) * 100).toFixed(2)) }}%</template>
                      </n-statistic>
                    </n-gi>
                  </n-grid>
                </n-card>

                <n-space justify="space-between" align="center">
                  <n-text depth="3">{{ t('adminSettings.serviceHealthSnapshot') }}</n-text>
                  <n-space>
                    <TableColumnSelector
                      v-model="serviceSelectedColumnKeys"
                      :options="serviceColumnOptions"
                      :visible-count="serviceVisibleColumnCount"
                      :total-count="serviceTotalColumnCount"
                      :button-label="t('common.showFields')"
                      :title="t('common.visibleFields')"
                      :hint="t('common.columnVisibilityHint')"
                      :reset-label="t('common.restoreDefaultFields')"
                      @reset="resetServiceSelectedColumns"
                    />
                    <n-text depth="3" style="font-size: 12px">{{ t('adminSettings.lastRefreshed') }}: {{ serverMonitoringGeneratedAt || '-' }}</n-text>
                  </n-space>
                </n-space>

                <n-data-table
                  :columns="serviceVisibleColumns"
                  :data="serviceStatusRows"
                  :pagination="false"
                  :bordered="false"
                  :scroll-x="serviceTableScrollX"
                />
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="debug" :tab="t('adminSettings.debugTools')">
              <n-space vertical :size="16">
                <n-card :title="t('adminSettings.systemOverview')" size="small">
                  <template #header-extra>
                    <n-space>
                      <n-button size="small" :type="debugAutoRefresh ? 'primary' : 'default'" @click="toggleDebugAutoRefresh(!debugAutoRefresh)">
                        {{ debugAutoRefresh ? t('adminSettings.stopRefresh') : t('adminSettings.autoRefresh') }}
                      </n-button>
                      <n-button size="small" :loading="loadingDebugStats" @click="loadDebugStats">{{ t('adminSettings.refresh') }}</n-button>
                      <n-button size="small" type="warning" @click="handleForceGC">{{ t('adminSettings.forceGC') }}</n-button>
                    </n-space>
                  </template>

                  <n-grid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.processResources')">
                        <n-space vertical size="small">
                          <div>
                            <n-space justify="space-between"><n-text>{{ t('adminSettings.cpu') }}</n-text><n-text>{{ Number((serverMonitoringData?.process.process_cpu || 0).toFixed(1)) }}%</n-text></n-space>
                            <n-progress type="line" :percentage="Number((serverMonitoringData?.process.process_cpu || 0).toFixed(1))" :status="(serverMonitoringData?.process.process_cpu ?? 0) > 80 ? 'error' : 'success'" :show-indicator="false" style="margin-top: 4px" />
                          </div>
                          <n-space justify="space-between"><n-text>{{ t('adminSettings.memory') }}</n-text><n-text>{{ formatStorageFromMB(serverMonitoringData?.process.process_rss_mb || 0) }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>{{ t('adminSettings.goroutines') }}</n-text><n-text>{{ serverMonitoringData?.process.goroutines || 0 }}</n-text></n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.goroutineStats')">
                        <n-space vertical size="small">
                          <n-space justify="space-between"><n-text>{{ t('adminSettings.runtimeTotal') }}</n-text><n-text>{{ debugStats?.total_count || 0 }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>{{ t('adminSettings.tracked') }}</n-text><n-text>{{ debugStats?.tracked_count || 0 }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>{{ t('adminSettings.potentialLeaks') }}</n-text><n-text type="error">{{ debugStats?.potential_leaks || 0 }}</n-text></n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                  </n-grid>
                </n-card>

                <n-card :title="t('adminSettings.pprofTitle')" size="small">
                  <template #header-extra>
                    <n-button size="small" @click="clearAllPprofResults">{{ t('adminSettings.clearResults') }}</n-button>
                  </template>

                  <n-grid :x-gap="12" :y-gap="12" cols="1 s:2 m:3 l:3" responsive="screen">
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.cpuProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.cpuProfileDesc') }}</n-text>
                          <n-space>
                            <n-input-number v-model:value="pprofConfig.cpuSeconds" :min="5" :max="120" size="small" style="width: 90px" />
                            <n-button size="small" type="primary" :loading="pprofLoading.cpu" @click="captureCPUProfile">{{ t('adminSettings.capture') }}</n-button>
                          </n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.heapProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.heapProfileDesc') }}</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.heap" @click="captureHeapProfile">{{ t('adminSettings.capture') }}</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.goroutineProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.goroutineProfileDesc') }}</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.goroutine" @click="captureGoroutineProfile">{{ t('adminSettings.capture') }}</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.allocsProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.allocsProfileDesc') }}</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.allocs" @click="captureAllocsProfile">{{ t('adminSettings.capture') }}</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.blockProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.blockProfileDesc') }}</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.block" @click="captureBlockProfile">{{ t('adminSettings.capture') }}</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" :title="t('adminSettings.mutexProfile')">
                        <n-space vertical size="small">
                          <n-text depth="3">{{ t('adminSettings.mutexProfileDesc') }}</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.mutex" @click="captureMutexProfile">{{ t('adminSettings.capture') }}</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                  </n-grid>

                  <n-empty v-if="!hasAnyPprofResult" :description="t('adminSettings.clickToCapture')" style="margin-top: 16px" />
                  <n-space v-else vertical :size="12" style="margin-top: 16px">
                    <n-card v-if="pprofResults.cpu" size="small" :title="t('adminSettings.cpuProfileResult')">
                      <n-code :code="pprofResults.cpuText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.heap" size="small" :title="t('adminSettings.heapProfileResult')">
                      <n-code :code="pprofResults.heapText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.goroutine" size="small" :title="t('adminSettings.goroutineProfileResult')">
                      <n-code :code="pprofResults.goroutine || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.allocs" size="small" :title="t('adminSettings.allocsProfileResult')">
                      <n-code :code="pprofResults.allocsText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.block" size="small" :title="t('adminSettings.blockProfileResult')">
                      <n-code :code="pprofResults.blockText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.mutex" size="small" :title="t('adminSettings.mutexProfileResult')">
                      <n-code :code="pprofResults.mutexText || ''" language="text" word-wrap />
                    </n-card>
                  </n-space>
                </n-card>

                <n-card :title="t('adminSettings.runtimeStacks')" size="small">
                  <template #header-extra>
                    <n-space>
                      <n-tooltip trigger="hover">
                        <template #trigger>
                          <n-input-number v-model:value="stackFilterMinWaitMinutes" :min="0" size="small" style="width: 140px" :placeholder="t('adminSettings.minWaitMinutes')" />
                        </template>
                        {{ t('adminSettings.filterTooltip') }}
                      </n-tooltip>
                      <n-button size="small" :loading="loadingRuntimeStacks" @click="loadRuntimeStacks">{{ t('adminSettings.loadStacks') }}</n-button>
                      <n-button size="small" @click="clearRuntimeStacks">{{ t('adminSettings.clearStacks') }}</n-button>
                    </n-space>
                  </template>
                  <n-empty v-if="runtimeStackText === ''" :description="t('adminSettings.clickToLoadStacks')" />
                  <n-code v-else :code="runtimeStackText" language="text" word-wrap />
                </n-card>
              </n-space>
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>
      </n-tabs>
    </n-spin>

    <n-modal v-model:show="showAddModal" preset="card" :title="t('adminSettings.addConfigTitle')" style="width: 500px;" :mask-closable="false">
      <n-form ref="addFormRef" :model="addForm" :rules="addFormRules" label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')" path="key">
          <n-input v-model:value="addForm.key" :placeholder="t('adminSettings.configKeyPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')" path="value">
          <n-input v-model:value="addForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')" path="label">
          <n-input v-model:value="addForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')" path="type">
          <n-select v-model:value="addForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')" path="description">
          <n-input v-model:value="addForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="addForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="adding" @click="handleAddSetting">{{ t('adminSettings.add') }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showEditModal" preset="card" :title="t('adminSettings.editConfigTitle')" style="width: 520px;" :mask-closable="false">
      <n-form label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')">
          <n-input v-model:value="editForm.key" disabled />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')">
          <n-input v-model:value="editForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')">
          <n-input v-model:value="editForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')">
          <n-select v-model:value="editForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')">
          <n-input v-model:value="editForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="editForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="savingEdit" @click="handleSaveSettingEdit">{{ t('common.save') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { computed, h, onMounted, onUnmounted, reactive, ref } from 'vue'
  import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
  import { useTableColumnVisibility } from '@/hooks'
  import {
    NAlert,
    NButton,
    NCode,
    NDataTable,
    NDescriptions,
    NDescriptionsItem,
    NDivider,
    NForm,
    NFormItem,
    NGi,
    NGrid,
    NInput,
    NInputNumber,
    NModal,
    NProgress,
    NSelect,
    NSpace,
    NSpin,
    NStatistic,
    NSwitch,
    NTabPane,
    NTabs,
    NTag,
    NText,
    NTooltip,
    type DataTableColumns,
    useMessage,
  } from 'naive-ui'
  import { adminApi } from '@/service/api/admin'
  import { adminDebugApi } from '@/service/api/admin/debug'
  import EmailTemplates from '@/views/admin/email-templates/index.vue'
  import SMSLogs from '@/views/admin/sms-logs/index.vue'
  import OperationLogs from '@/views/admin/logs/index.vue'
  import type { ServerMonitoringStatusResponse, SettingDTO, SettingType } from '@/service/api/admin/settings'
  import { useSettingsStore } from '@/store/settings'
  // authStorage 读取当前活跃作用域 token，避免双窗口/隔离登录时误用 local 中的用户态 token
  import { authStorage, parseBooleanSetting } from '@/utils'

  const message = useMessage()
  const settingsStore = useSettingsStore()
  const { t } = useI18n()

  const loading = ref(true)
  const adding = ref(false)
  const savingEdit = ref(false)
  const showAddModal = ref(false)
  const showEditModal = ref(false)
  const savingBasic = ref(false)
  const savingEmail = ref(false)
  const savingSecurity = ref(false)
  const savingRealnameApi = ref(false)
  const savingPayment = ref(false)
  const testingEmail = ref(false)
  const restartingBackend = ref(false)
  const loadingServerMonitoring = ref(false)
  const topTab = ref('system-config')
  const systemSubTab = ref('basic')
  const mode = import.meta.env.MODE
  const buildTime = typeof __BUILD_TIMESTAMP__ !== 'undefined' ? __BUILD_TIMESTAMP__ : t('adminSettings.developmentMode')
  const serverMonitoringGeneratedAt = ref('')
  const serverMonitoringData = ref<ServerMonitoringStatusResponse | null>(null)

  const loadingDebugStats = ref(false)
  const debugAutoRefresh = ref(false)
  const debugRefreshInterval = ref<number | null>(null)
  const debugStats = ref<any>(null)
  const loadingRuntimeStacks = ref(false)
  const runtimeStackText = ref('')
  const stackFilterMinWaitMinutes = ref(0)

  const pprofConfig = ref({
    cpuSeconds: 30,
  })

  const pprofLoading = reactive({
    cpu: false,
    heap: false,
    goroutine: false,
    allocs: false,
    block: false,
    mutex: false,
  })

  const pprofResults = ref({
    cpu: false,
    cpuText: '',
    heap: false,
    heapText: '',
    heapStats: null as { alloc: number, objects: number } | null,
    goroutine: '',
    goroutineCount: 0,
    allocs: false,
    allocsText: '',
    block: false,
    blockText: '',
    mutex: false,
    mutexText: '',
  })

  const hasAnyPprofResult = computed(() => {
    const r = pprofResults.value
    return r.cpu || r.heap || !!r.goroutine || r.allocs || r.block || r.mutex
  })

  const switchLoading = reactive({
    allow_register: false,
    allow_delete_account: false,
    smtp_ssl: false,
    geetest_enabled: false,
    realname_enabled: false,
    realname_review_required: false,
    realname_api_enabled: false,
    email_verify_enabled: false,
    sms_verify_enabled: false,
    payment_enabled: false,
    withdraw_enabled: false,
  })

  const langOptions = [
    { label: t('adminSettings.langZhCN'), value: 'zhCN' },
    { label: t('adminSettings.langEnUS'), value: 'enUS' },
  ]

const smsProviderOptions = [
  { label: t('adminSettings.smsProviderConsole'), value: 'console' },
  { label: t('adminSettings.smsProviderAliyun'), value: 'aliyun' },
  { label: t('adminSettings.smsProviderTencent'), value: 'tencent' },
  { label: t('adminSettings.smsProviderCustom'), value: 'custom' },
]

const smsBodyFormatOptions = [
  { label: t('adminSettings.formatJSON'), value: 'json' },
  { label: t('adminSettings.formatForm'), value: 'form' },
]

const realnameApiProviderOptions = [
  { label: t('adminSettings.providerAliyun'), value: 'aliyun' },
  { label: t('adminSettings.providerTencent'), value: 'tencent' },
  { label: t('adminSettings.providerBaidu'), value: 'baidu' },
  { label: t('adminSettings.providerCustom'), value: 'custom' },
]

const typeOptions = [
  { label: t('adminSettings.typeString'), value: 'string' },
  { label: t('adminSettings.typeNumber'), value: 'number' },
  { label: t('adminSettings.typeBoolean'), value: 'boolean' },
  { label: t('adminSettings.typeJSON'), value: 'json' },
]

const basicForm = reactive({
  site_name: '',
  site_desc: '',
  site_logo: '',
  copyright: '',
  icp: '',
  version: '',
  default_lang: 'zhCN',
  allow_register: true,
  allow_delete_account: false,
  frontend_url: '',
  backend_api_url: '',
})

const emailForm = reactive({
  email_verify_enabled: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_ssl: true,
  system_email_name: '',
})

const smsForm = reactive({
  sms_verify_enabled: false,
  sms_provider: 'console',
  sms_access_key: '',
  sms_secret_key: '',
  sms_sign_name: '',
  sms_template_code: '',
  sms_template_code_en: '',
  sms_region: '',
  sms_sdk_app_id: '',
  sms_endpoint: '',
  sms_body_format: 'json',
})

const savingSms = ref(false)

const smsProviderNeedsSignName = computed(() => smsForm.sms_provider !== 'console')
const smsProviderNeedsTemplateCode = computed(() => ['aliyun', 'tencent'].includes(smsForm.sms_provider))
const smsAccessKeyPlaceholder = computed(() => {
  if (smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretId')
  if (smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiKeyOptional')
  return t('adminSettings.smsProviderAccessKey')
})
const smsSecretKeyPlaceholder = computed(() => {
  if (smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretKey')
  if (smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiSecretOptional')
  return t('adminSettings.smsProviderSecretKey')
})
const smsTemplateLabel = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCode') : t('adminSettings.smsTemplateId'))
const smsTemplatePlaceholder = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeExample') : t('adminSettings.smsTemplateIdPlaceholder'))
const smsTemplateEnLabel = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglish') : t('adminSettings.smsTemplateIdEnglish'))
const smsTemplateEnPlaceholder = computed(() => smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglishOptional') : t('adminSettings.smsTemplateIdEnglishOptional'))

const securityForm = reactive({
  geetest_enabled: false,
  geetest_captcha_id: '',
  geetest_captcha_key: '',
  jwt_access_expire: 7200,
  jwt_refresh_expire: 604800,
  login_max_failure: 5,
  login_lock_duration: 10,
  allow_delete_account: false,
  realname_enabled: true,
  realname_review_required: true,
  realname_notify_text: t('adminSettings.realnameNotifyTextDefault'),
})

const realnameApiForm = reactive({
  realname_api_enabled: false,
  realname_api_provider: 'aliyun',
  realname_api_app_key: '',
  realname_api_app_secret: '',
  realname_api_endpoint: '',
})

const paymentForm = reactive({
  payment_enabled: false,
  payment_order_expire_minutes: 30,
  withdraw_enabled: true,
  withdraw_min_amount: 10,
  withdraw_notify_text: t('adminSettings.withdrawNotifyTextDefault'),
  withdraw_account_types_text: '["bank","alipay","wechat","usdt"]',
})

const customSettings = ref<SettingDTO[]>([])

const addFormRef = ref()
const addForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as string,
  description: '',
  is_public: false,
})

const addFormRules = {
  key: [
    { required: true, message: () => t('adminSettings.keyRequired'), trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: () => t('adminSettings.keyPattern'), trigger: 'blur' },
  ],
  label: [{ required: true, message: () => t('adminSettings.labelRequired'), trigger: 'blur' }],
}

const editForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as SettingType,
  description: '',
  is_public: false,
})

const plugins = ref([
  { name: 'Demo Plugin', version: '1.0.0', active: true },
  { name: 'Email', version: '1.0.0', active: true },
])

const customColumns: DataTableColumns<SettingDTO> = [
  { title: t('adminSettings.columnKey'), key: 'key' },
  { title: t('adminSettings.columnLabel'), key: 'label' },
  { title: t('adminSettings.columnValue'), key: 'value', ellipsis: { tooltip: true } },
  { title: t('adminSettings.columnType'), key: 'type', width: 80 },
  {
    title: t('adminSettings.columnPublic'),
    key: 'is_public',
    width: 80,
    render: row => row.is_public ? t('adminSettings.yes') : t('adminSettings.no'),
  },
  {
    title: t('adminSettings.columnActions'),
    key: 'actions',
    width: 180,
    render: (row) => {
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEditSetting(row),
          }, () => t('adminSettings.edit')),
          h(NButton, {
            size: 'small',
            type: 'error',
            text: true,
            onClick: () => handleDeleteSetting(row.key),
          }, () => t('adminSettings.delete')),
        ],
      })
    },
  },
]

const customSelectableColumnOptions = computed(() => [
  { key: 'key', label: t('adminSettings.columnKey') },
  { key: 'label', label: t('adminSettings.columnLabel') },
  { key: 'value', label: t('adminSettings.columnValue') },
  { key: 'type', label: t('adminSettings.columnType') },
  { key: 'is_public', label: t('adminSettings.columnPublic') },
])

const {
  columnOptions: customColumnOptions,
  selectedColumnKeys: customSelectedColumnKeys,
  visibleColumns: customVisibleColumns,
  visibleColumnCount: customVisibleColumnCount,
  totalColumnCount: customTotalColumnCount,
  tableScrollX: customTableScrollX,
  resetSelectedColumns: resetCustomSelectedColumns,
} = useTableColumnVisibility<SettingDTO>({
  storageKey: 'admin-settings-custom-list',
  columns: customColumns,
  options: customSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})

type ServiceStatusRow = {
  name: string
  status: 'up' | 'down' | 'warning'
  message: string
  detail: string
}

const serviceStatusRows = computed<ServiceStatusRow[]>(() => {
  if (!serverMonitoringData.value?.services) {
    return []
  }
  return serverMonitoringData.value.services.map(service => {
    const detailParts: string[] = []
    if (typeof service.open_connections === 'number') {
      detailParts.push(t('adminSettings.serviceConnections', { count: service.open_connections }))
    }
    if (typeof service.in_use === 'number') {
      detailParts.push(t('adminSettings.serviceInUse', { count: service.in_use }))
    }
    if (typeof service.idle === 'number') {
      detailParts.push(t('adminSettings.serviceIdle', { count: service.idle }))
    }
    if (service.host && service.port) {
      detailParts.push(`${service.host}:${service.port}`)
    }
    return {
      name: service.name,
      status: service.status,
      message: service.message,
      detail: detailParts.join(' | '),
    }
  })
})

const serviceStatusColumns: DataTableColumns<ServiceStatusRow> = [
  { title: t('adminSettings.columnService'), key: 'name' },
  {
    title: t('adminSettings.columnStatus'),
    key: 'status',
    width: 100,
    render: row => {
      const type = row.status === 'up' ? 'success' : row.status === 'warning' ? 'warning' : 'error'
      const text = row.status === 'up' ? t('adminSettings.statusNormal') : row.status === 'warning' ? t('adminSettings.statusWarning') : t('adminSettings.statusError')
      return h(NTag, { type, size: 'small' }, () => text)
    },
  },
  { title: t('adminSettings.columnMessage'), key: 'message' },
  { title: t('adminSettings.columnDetail'), key: 'detail' },
]

const serviceSelectableColumnOptions = computed(() => [
  { key: 'name', label: t('adminSettings.columnService') },
  { key: 'status', label: t('adminSettings.columnStatus') },
  { key: 'message', label: t('adminSettings.columnMessage') },
  { key: 'detail', label: t('adminSettings.columnDetail') },
])

const {
  columnOptions: serviceColumnOptions,
  selectedColumnKeys: serviceSelectedColumnKeys,
  visibleColumns: serviceVisibleColumns,
  visibleColumnCount: serviceVisibleColumnCount,
  totalColumnCount: serviceTotalColumnCount,
  tableScrollX: serviceTableScrollX,
  resetSelectedColumns: resetServiceSelectedColumns,
} = useTableColumnVisibility<ServiceStatusRow>({
  storageKey: 'admin-settings-service-health-list',
  columns: serviceStatusColumns,
  options: serviceSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})

const cpuPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.cpu.usage_percent ?? 0))
const memoryPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.memory.used_percent ?? 0))
const swapPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.swap.used_percent ?? 0))
const diskPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.disk.used_percent ?? 0))

function normalizePercent(value: number): number {
  if (!Number.isFinite(value)) {
    return 0
  }
  if (value < 0) {
    return 0
  }
  if (value > 100) {
    return 100
  }
  return Number(value.toFixed(2))
}

function formatPercent(value: number): string {
  return `${normalizePercent(value).toFixed(2)}%`
}

function formatInteger(value: number): string {
  if (!Number.isFinite(value)) {
    return '-'
  }
  return Math.round(value).toLocaleString()
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx++
  }
  return `${size.toFixed(2)} ${units[idx]}`
}

function formatStorageFromMB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const gb = value / 1024
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(2)} TB`
  }
  if (gb >= 1) {
    return `${gb.toFixed(2)} GB`
  }
  return `${value.toFixed(2)} MB`
}

function formatStorageFromGB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} TB`
  }
  return `${value.toFixed(2)} GB`
}

function formatGeneratedAt(value: string | undefined) {
  if (!value) {
    return ''
  }
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) {
    return value
  }
  return d.toLocaleString()
}

function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '-'
  }
  const day = Math.floor(seconds / 86400)
  const hour = Math.floor((seconds % 86400) / 3600)
  const minute = Math.floor((seconds % 3600) / 60)
  return t('adminSettings.uptimeFormat', { day, hour, minute })
}

function formatUptimePrecise(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '-'
  }
  const day = Math.floor(seconds / 86400)
  const hour = Math.floor((seconds % 86400) / 3600)
  const minute = Math.floor((seconds % 3600) / 60)
  const second = Math.floor(seconds % 60)
  return t('adminSettings.uptimePreciseFormat', { day, hour, minute, second })
}

function formatStartTimeFromUptime(generatedAt?: string, uptimeSeconds?: number): string {
  if (!generatedAt || !Number.isFinite(uptimeSeconds || NaN)) {
    return '-'
  }
  const generated = new Date(generatedAt)
  if (Number.isNaN(generated.getTime())) {
    return '-'
  }
  const start = new Date(generated.getTime() - (uptimeSeconds || 0) * 1000)
  const mm = String(start.getMonth() + 1).padStart(2, '0')
  const dd = String(start.getDate()).padStart(2, '0')
  const hh = String(start.getHours()).padStart(2, '0')
  const mi = String(start.getMinutes()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${mi}`
}

const uptimeText = computed(() => formatUptime(serverMonitoringData.value?.uptime_seconds ?? 0))
const uptimeTextPrecise = computed(() => formatUptimePrecise(serverMonitoringData.value?.uptime_seconds ?? 0))
const startTimeText = computed(() => formatStartTimeFromUptime(serverMonitoringData.value?.generated_at, serverMonitoringData.value?.uptime_seconds))

async function loadServerMonitoringStatus() {
  loadingServerMonitoring.value = true
  try {
    const response = await adminApi.settings.serverMonitoring()
    serverMonitoringData.value = response.data ?? null
    serverMonitoringGeneratedAt.value = formatGeneratedAt(response.data?.generated_at)
  }
  catch (error: any) {
    serverMonitoringData.value = null
    serverMonitoringGeneratedAt.value = ''
    message.error(t('adminSettings.loadMonitoringFailed') + (error.message || ''))
  }
  finally {
    loadingServerMonitoring.value = false
  }
}

async function loadSettings() {
  loading.value = true
  try {
    const response = await adminApi.settings.list()
    if (response.data?.categories) {
      for (const category of response.data.categories) {
        for (const item of category.items) {
          if (item.key === 'site_name') basicForm.site_name = String(item.value || '')
          if (item.key === 'site_desc') basicForm.site_desc = String(item.value || '')
          if (item.key === 'site_logo') basicForm.site_logo = String(item.value || '')
          if (item.key === 'copyright') basicForm.copyright = String(item.value || '')
          if (item.key === 'icp') basicForm.icp = String(item.value || '')
          if (item.key === 'version') basicForm.version = String(item.value || '')
          if (item.key === 'default_lang') basicForm.default_lang = String(item.value || 'zhCN')

          if (item.key === 'allow_register') basicForm.allow_register = parseBooleanSetting(item.value)
          if (item.key === 'allow_delete_account') securityForm.allow_delete_account = parseBooleanSetting(item.value)
          if (item.key === 'frontend_url') basicForm.frontend_url = String(item.value || '')
          if (item.key === 'backend_api_url') basicForm.backend_api_url = String(item.value || '')

          if (item.key === 'email_verify_enabled') emailForm.email_verify_enabled = parseBooleanSetting(item.value)
          if (item.key === 'smtp_host') emailForm.smtp_host = String(item.value || '')
          if (item.key === 'smtp_port') emailForm.smtp_port = Number(item.value) || 587
          if (item.key === 'smtp_username') emailForm.smtp_username = String(item.value || '')
          if (item.key === 'smtp_password') emailForm.smtp_password = String(item.value || '')
          if (item.key === 'smtp_ssl') emailForm.smtp_ssl = parseBooleanSetting(item.value)
          if (item.key === 'system_email_name') emailForm.system_email_name = String(item.value || '')

          if (item.key === 'sms_verify_enabled') smsForm.sms_verify_enabled = parseBooleanSetting(item.value)
          if (item.key === 'sms_provider') smsForm.sms_provider = String(item.value || 'console')
          if (item.key === 'sms_access_key') smsForm.sms_access_key = String(item.value || '')
          if (item.key === 'sms_secret_key') smsForm.sms_secret_key = String(item.value || '')
          if (item.key === 'sms_sign_name') smsForm.sms_sign_name = String(item.value || '')
          if (item.key === 'sms_template_code') smsForm.sms_template_code = String(item.value || '')
          if (item.key === 'sms_template_code_en') smsForm.sms_template_code_en = String(item.value || '')
          if (item.key === 'sms_region') smsForm.sms_region = String(item.value || '')
          if (item.key === 'sms_sdk_app_id') smsForm.sms_sdk_app_id = String(item.value || '')
          if (item.key === 'sms_endpoint') smsForm.sms_endpoint = String(item.value || '')
          if (item.key === 'sms_body_format') smsForm.sms_body_format = String(item.value || 'json')

          if (item.key === 'geetest_enabled') securityForm.geetest_enabled = parseBooleanSetting(item.value)
          if (item.key === 'geetest_captcha_id') securityForm.geetest_captcha_id = String(item.value || '')
          if (item.key === 'geetest_captcha_key') securityForm.geetest_captcha_key = String(item.value || '')
          if (item.key === 'jwt_access_expire') securityForm.jwt_access_expire = Number(item.value) || 7200
          if (item.key === 'jwt_refresh_expire') securityForm.jwt_refresh_expire = Number(item.value) || 604800
          if (item.key === 'login_max_failure') securityForm.login_max_failure = Number(item.value) || 5
          if (item.key === 'login_lock_duration') securityForm.login_lock_duration = Number(item.value) || 10
          if (item.key === 'realname_enabled') securityForm.realname_enabled = parseBooleanSetting(item.value)
          if (item.key === 'realname_review_required') securityForm.realname_review_required = parseBooleanSetting(item.value)
          if (item.key === 'realname_notify_text') securityForm.realname_notify_text = String(item.value || t('adminSettings.realnameNotifyTextDefault'))

          if (item.key === 'realname_api_enabled') realnameApiForm.realname_api_enabled = parseBooleanSetting(item.value)
          if (item.key === 'realname_api_provider') realnameApiForm.realname_api_provider = String(item.value || 'aliyun')
          if (item.key === 'realname_api_app_key') realnameApiForm.realname_api_app_key = String(item.value || '')
          if (item.key === 'realname_api_app_secret') realnameApiForm.realname_api_app_secret = String(item.value || '')
          if (item.key === 'realname_api_endpoint') realnameApiForm.realname_api_endpoint = String(item.value || '')

          if (item.key === 'payment_enabled') paymentForm.payment_enabled = parseBooleanSetting(item.value)
          if (item.key === 'payment_order_expire_minutes') paymentForm.payment_order_expire_minutes = Number(item.value) || 30
          if (item.key === 'withdraw_enabled') paymentForm.withdraw_enabled = parseBooleanSetting(item.value)
          if (item.key === 'withdraw_min_amount') paymentForm.withdraw_min_amount = Number(item.value) || 10
          if (item.key === 'withdraw_notify_text') paymentForm.withdraw_notify_text = String(item.value || t('adminSettings.withdrawNotifyTextDefault'))
          if (item.key === 'withdraw_account_types') paymentForm.withdraw_account_types_text = typeof item.value === 'string' ? item.value : JSON.stringify(item.value || ['bank', 'alipay', 'wechat', 'usdt'])
        }

        if (category.category === 'custom') {
          customSettings.value = category.items
        }
      }
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.loadSettingsFailed'))
  }
  finally {
    loading.value = false
  }
}

async function handleUpdateAllowRegister(nextValue: boolean) {
  const prev = basicForm.allow_register
  basicForm.allow_register = nextValue
  switchLoading.allow_register = true
  try {
    const res = await adminApi.settings.update('allow_register', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ allow_register: nextValue })
      message.success(res.message || t('adminSettings.registerSwitchUpdated'))
    }
    else {
      basicForm.allow_register = prev
      message.error(res.message || t('adminSettings.registerSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    basicForm.allow_register = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.allow_register = false
  }
}

async function handleUpdateAllowDeleteAccount(nextValue: boolean) {
  const prev = securityForm.allow_delete_account
  securityForm.allow_delete_account = nextValue
  switchLoading.allow_delete_account = true
  try {
    const res = await adminApi.settings.update('allow_delete_account', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ allow_delete_account: nextValue })
      message.success(res.message || t('adminSettings.deleteAccountSwitchUpdated'))
    }
    else {
      securityForm.allow_delete_account = prev
      message.error(res.message || t('adminSettings.deleteAccountSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.allow_delete_account = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.allow_delete_account = false
  }
}

async function handleUpdateSmtpSSL(nextValue: boolean) {
  const prev = emailForm.smtp_ssl
  emailForm.smtp_ssl = nextValue
  switchLoading.smtp_ssl = true
  try {
    const res = await adminApi.settings.update('smtp_ssl', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.smtpSslUpdated'))
    }
    else {
      emailForm.smtp_ssl = prev
      message.error(res.message || t('adminSettings.smtpSslUpdateFailed'))
    }
  }
  catch (error: any) {
    emailForm.smtp_ssl = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.smtp_ssl = false
  }
}

async function handleUpdateEmailVerifyEnabled(nextValue: boolean) {
  const prev = emailForm.email_verify_enabled
  emailForm.email_verify_enabled = nextValue
  switchLoading.email_verify_enabled = true
  try {
    const res = await adminApi.settings.update('email_verify_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ email_verify_enabled: nextValue })
      message.success(res.message || t('adminSettings.emailVerifySwitchUpdated'))
    }
    else {
      emailForm.email_verify_enabled = prev
      message.error(res.message || t('adminSettings.emailVerifySwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    emailForm.email_verify_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.email_verify_enabled = false
  }
}

async function handleUpdateSmsVerifyEnabled(nextValue: boolean) {
  const prev = smsForm.sms_verify_enabled
  smsForm.sms_verify_enabled = nextValue
  switchLoading.sms_verify_enabled = true
  try {
    const res = await adminApi.settings.update('sms_verify_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ sms_verify_enabled: nextValue })
      message.success(res.message || t('adminSettings.smsVerifySwitchUpdated'))
    }
    else {
      smsForm.sms_verify_enabled = prev
      message.error(res.message || t('adminSettings.smsVerifySwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    smsForm.sms_verify_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.sms_verify_enabled = false
  }
}

async function handleSaveSms() {
  savingSms.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      sms_provider: smsForm.sms_provider,
      sms_access_key: smsForm.sms_access_key,
      sms_secret_key: smsForm.sms_secret_key,
      sms_sign_name: smsForm.sms_sign_name,
      sms_template_code: smsForm.sms_template_code,
      sms_template_code_en: smsForm.sms_template_code_en,
      sms_region: smsForm.sms_region,
      sms_sdk_app_id: smsForm.sms_sdk_app_id,
      sms_endpoint: smsForm.sms_endpoint,
      sms_body_format: smsForm.sms_body_format,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.smsSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.smsSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingSms.value = false
  }
}

async function handleUpdateGeetestEnabled(nextValue: boolean) {
  const prev = securityForm.geetest_enabled
  securityForm.geetest_enabled = nextValue
  switchLoading.geetest_enabled = true
  try {
    const res = await adminApi.settings.update('geetest_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ geetest_enabled: nextValue })
      message.success(res.message || t('adminSettings.geetestSwitchUpdated'))
    }
    else {
      securityForm.geetest_enabled = prev
      message.error(res.message || t('adminSettings.geetestSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.geetest_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.geetest_enabled = false
  }
}

async function handleUpdateRealnameEnabled(nextValue: boolean) {
  const prev = securityForm.realname_enabled
  securityForm.realname_enabled = nextValue
  switchLoading.realname_enabled = true
  try {
    const res = await adminApi.settings.update('realname_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameSwitchUpdated'))
    }
    else {
      securityForm.realname_enabled = prev
      message.error(res.message || t('adminSettings.realnameSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.realname_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_enabled = false
  }
}

async function handleUpdateRealnameReviewRequired(nextValue: boolean) {
  const prev = securityForm.realname_review_required
  securityForm.realname_review_required = nextValue
  switchLoading.realname_review_required = true
  try {
    const res = await adminApi.settings.update('realname_review_required', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameReviewSwitchUpdated'))
    }
    else {
      securityForm.realname_review_required = prev
      message.error(res.message || t('adminSettings.realnameReviewSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    securityForm.realname_review_required = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_review_required = false
  }
}

async function handleUpdateRealnameApiEnabled(nextValue: boolean) {
  const prev = realnameApiForm.realname_api_enabled
  realnameApiForm.realname_api_enabled = nextValue
  switchLoading.realname_api_enabled = true
  try {
    const res = await adminApi.settings.update('realname_api_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameApiSwitchUpdated'))
    }
    else {
      realnameApiForm.realname_api_enabled = prev
      message.error(res.message || t('adminSettings.realnameApiSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    realnameApiForm.realname_api_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.realname_api_enabled = false
  }
}

async function handleSaveRealnameApi() {
  savingRealnameApi.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      realname_api_provider: realnameApiForm.realname_api_provider,
      realname_api_app_key: realnameApiForm.realname_api_app_key,
      realname_api_app_secret: realnameApiForm.realname_api_app_secret,
      realname_api_endpoint: realnameApiForm.realname_api_endpoint,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.realnameApiSaveSuccess'))
    }
    else {
      message.error(res.message || t('adminSettings.realnameApiSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingRealnameApi.value = false
  }
}

async function handleUpdatePaymentEnabled(nextValue: boolean) {
  const prev = paymentForm.payment_enabled
  paymentForm.payment_enabled = nextValue
  switchLoading.payment_enabled = true
  try {
    const res = await adminApi.settings.update('payment_enabled', String(nextValue))
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.paymentSwitchUpdated'))
    }
    else {
      paymentForm.payment_enabled = prev
      message.error(res.message || t('adminSettings.paymentSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    paymentForm.payment_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.payment_enabled = false
  }
}

async function handleUpdateWithdrawEnabled(nextValue: boolean) {
  const prev = paymentForm.withdraw_enabled
  paymentForm.withdraw_enabled = nextValue
  switchLoading.withdraw_enabled = true
  try {
    const res = await adminApi.settings.update('withdraw_enabled', String(nextValue))
    if (res.isSuccess) {
      settingsStore.updateConfig({ withdraw_enabled: nextValue })
      message.success(res.message || t('adminSettings.withdrawSwitchUpdated'))
    }
    else {
      paymentForm.withdraw_enabled = prev
      message.error(res.message || t('adminSettings.withdrawSwitchUpdateFailed'))
    }
  }
  catch (error: any) {
    paymentForm.withdraw_enabled = prev
    message.error(t('adminSettings.updateFailed'))
  }
  finally {
    switchLoading.withdraw_enabled = false
  }
}

async function handleSavePayment() {
  savingPayment.value = true
  try {
    const parsedAccountTypes = JSON.parse(paymentForm.withdraw_account_types_text || '[]')
    if (!Array.isArray(parsedAccountTypes) || parsedAccountTypes.length === 0 || parsedAccountTypes.some(item => typeof item !== 'string' || !item.trim())) {
      throw new Error(t('adminSettings.invalidAccountTypes'))
    }
    const res = await adminApi.settings.batchUpdate({
      payment_order_expire_minutes: String(paymentForm.payment_order_expire_minutes),
      withdraw_min_amount: String(paymentForm.withdraw_min_amount),
      withdraw_notify_text: paymentForm.withdraw_notify_text,
      withdraw_account_types: paymentForm.withdraw_account_types_text,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        withdraw_min_amount: paymentForm.withdraw_min_amount,
        withdraw_notify_text: paymentForm.withdraw_notify_text,
        withdraw_account_types: parsedAccountTypes,
      })
      message.success(res.message || t('adminSettings.paymentSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.paymentSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingPayment.value = false
  }
}

async function handleSaveBasic() {
  savingBasic.value = true
  try {
    const frontendUrl = basicForm.frontend_url.trim().replace(/\/+$/, '')
    const backendApiUrl = basicForm.backend_api_url.trim().replace(/\/+$/, '')
    basicForm.frontend_url = frontendUrl
    basicForm.backend_api_url = backendApiUrl
    const res = await adminApi.settings.batchUpdate({
      site_name: basicForm.site_name,
      site_desc: basicForm.site_desc,
      site_logo: basicForm.site_logo,
      copyright: basicForm.copyright,
      icp: basicForm.icp,
      version: basicForm.version,
      default_lang: basicForm.default_lang,
      frontend_url: frontendUrl,
      backend_api_url: backendApiUrl,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        site_name: basicForm.site_name,
        site_desc: basicForm.site_desc,
        site_logo: basicForm.site_logo,
        copyright: basicForm.copyright,
        icp: basicForm.icp,
        version: basicForm.version,
        default_lang: basicForm.default_lang,
      })
      message.success(res.message || t('adminSettings.basicSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.basicSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingBasic.value = false
  }
}

async function handleSaveEmail() {
  savingEmail.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      smtp_host: emailForm.smtp_host,
      smtp_port: String(emailForm.smtp_port),
      smtp_username: emailForm.smtp_username,
      smtp_password: emailForm.smtp_password,
      system_email_name: emailForm.system_email_name,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.emailSettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.emailSettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingEmail.value = false
  }
}

async function handleTestEmail() {
  testingEmail.value = true
  try {
    const res = await adminApi.emailTemplate.sendTest({
      to: emailForm.smtp_username,
    })
    if (res.isSuccess) {
      message.success(res.data?.message || t('adminSettings.testEmailSent'))
    }
    else {
      message.error(res.message || t('adminSettings.testEmailFailed'))
    }
  }
  catch (error: any) {
    message.error(error?.message || t('adminSettings.testEmailFailed'))
  }
  finally {
    testingEmail.value = false
  }
}

async function handleSaveSecurity() {
  savingSecurity.value = true
  try {
    const res = await adminApi.settings.batchUpdate({
      geetest_captcha_id: securityForm.geetest_captcha_id,
      geetest_captcha_key: securityForm.geetest_captcha_key,
      jwt_access_expire: String(securityForm.jwt_access_expire),
      jwt_refresh_expire: String(securityForm.jwt_refresh_expire),
      login_max_failure: String(securityForm.login_max_failure),
      login_lock_duration: String(securityForm.login_lock_duration),
      realname_notify_text: securityForm.realname_notify_text,
    })
    if (res.isSuccess) {
      settingsStore.updateConfig({
        geetest_captcha_id: securityForm.geetest_captcha_id,
      })
      message.success(res.message || t('adminSettings.securitySettingsSaved'))
    }
    else {
      message.error(res.message || t('adminSettings.securitySettingsSaveFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.saveFailed'))
  }
  finally {
    savingSecurity.value = false
  }
}

async function handleRestartBackend() {
  restartingBackend.value = true
  try {
    const res = await adminApi.settings.restartBackend()
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.restartBackendRequested'))
    }
    else {
      message.error(res.message || t('adminSettings.restartBackendFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.restartFailed'))
  }
  finally {
    restartingBackend.value = false
  }
}

async function handleAddSetting() {
  try {
    await addFormRef.value?.validate()
  }
  catch {
    return
  }

  adding.value = true
  try {
    const res = await adminApi.settings.create({
      key: addForm.key,
      value: addForm.value,
      type: addForm.type as SettingType,
      category: 'custom',
      label: addForm.label,
      description: addForm.description,
      is_public: addForm.is_public,
      is_editable: true,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemAdded'))
      showAddModal.value = false
      addForm.key = ''
      addForm.value = ''
      addForm.label = ''
      addForm.type = 'string'
      addForm.description = ''
      addForm.is_public = false
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemAddFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.addFailed'))
  }
  finally {
    adding.value = false
  }
}

async function handleDeleteSetting(key: string) {
  try {
    const res = await adminApi.settings.delete(key)
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemDeleted'))
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemDeleteFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.deleteFailed') + (error.message || ''))
  }
}

function handleEditSetting(row: SettingDTO) {
  editForm.key = row.key
  editForm.value = row.value == null ? '' : String(row.value)
  editForm.label = row.label || ''
  editForm.type = row.type
  editForm.description = row.description || ''
  editForm.is_public = Boolean(row.is_public)
  showEditModal.value = true
}

async function handleSaveSettingEdit() {
  if (!editForm.key) {
    return
  }
  savingEdit.value = true
  try {
    const res = await adminApi.settings.updateMeta(editForm.key, {
      value: editForm.value,
      type: editForm.type,
      category: 'custom',
      label: editForm.label,
      description: editForm.description,
      is_public: editForm.is_public,
      is_editable: true,
    })
    if (res.isSuccess) {
      message.success(res.message || t('adminSettings.configItemUpdated'))
      showEditModal.value = false
      await loadSettings()
    }
    else {
      message.error(res.message || t('adminSettings.configItemUpdateFailed'))
    }
  }
  catch (error: any) {
    message.error(t('adminSettings.editFailed'))
  }
  finally {
    savingEdit.value = false
  }
}

async function loadDebugStats() {
  loadingDebugStats.value = true
  try {
    const res = await adminDebugApi.goroutineStats()
    if (res.data) {
      debugStats.value = res.data
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.loadDebugStatsFailed') + e.message)
  }
  finally {
    loadingDebugStats.value = false
  }
}

function toggleDebugAutoRefresh(value: boolean) {
  debugAutoRefresh.value = value
  if (value) {
    debugRefreshInterval.value = window.setInterval(loadDebugStats, 3000)
  }
  else {
    if (debugRefreshInterval.value) {
      clearInterval(debugRefreshInterval.value)
      debugRefreshInterval.value = null
    }
  }
}

async function handleForceGC() {
  try {
    const res = await adminDebugApi.forceGC()
    if (res.data) {
      message.success(t('adminSettings.gcCompleted', { before: res.data.goroutines_before, after: res.data.goroutines_after }))
      loadDebugStats()
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.operationFailed') + e.message)
  }
}

function clearAllPprofResults() {
  pprofResults.value = {
    cpu: false,
    cpuText: '',
    heap: false,
    heapText: '',
    heapStats: null,
    goroutine: '',
    goroutineCount: 0,
    allocs: false,
    allocsText: '',
    block: false,
    blockText: '',
    mutex: false,
    mutexText: '',
  }
  message.success(t('adminSettings.resultsCleared'))
}

async function captureCPUProfile() {
  pprofLoading.cpu = true
  pprofResults.value.cpu = false
  pprofResults.value.cpuText = ''
  message.info(t('adminSettings.cpuProfileStarting', { seconds: pprofConfig.value.cpuSeconds }))
  try {
    const url = adminDebugApi.cpuProfile(pprofConfig.value.cpuSeconds)
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.cpu = true
      pprofResults.value.cpuText = text
      message.success(t('adminSettings.cpuProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.cpu = false
  }
}

async function captureHeapProfile() {
  pprofLoading.heap = true
  pprofResults.value.heap = false
  pprofResults.value.heapText = ''
  pprofResults.value.heapStats = null
  try {
    const url = adminDebugApi.heapProfile()
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.heap = true
      pprofResults.value.heapText = text
      const allocMatch = text.match(/# Alloc = (\d+)/)
      const objectsMatch = text.match(/# HeapObjects = (\d+)/)
      if (allocMatch || objectsMatch) {
        pprofResults.value.heapStats = {
          alloc: allocMatch ? Number(allocMatch[1]) : 0,
          objects: objectsMatch ? Number(objectsMatch[1]) : 0,
        }
      }
      message.success(t('adminSettings.heapProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.heap = false
  }
}

async function captureGoroutineProfile() {
  pprofLoading.goroutine = true
  pprofResults.value.goroutine = ''
  pprofResults.value.goroutineCount = 0
  try {
    const url = adminDebugApi.goroutineProfile(0)
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.goroutine = text
      const matches = text.match(/goroutine \d+/g)
      pprofResults.value.goroutineCount = matches ? matches.length : 0
      message.success(t('adminSettings.goroutineProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.goroutine = false
  }
}

async function captureAllocsProfile() {
  pprofLoading.allocs = true
  pprofResults.value.allocs = false
  pprofResults.value.allocsText = ''
  try {
    const url = adminDebugApi.allocsProfile()
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.allocs = true
      pprofResults.value.allocsText = text
      message.success(t('adminSettings.allocsProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.allocs = false
  }
}

async function captureBlockProfile() {
  pprofLoading.block = true
  pprofResults.value.block = false
  pprofResults.value.blockText = ''
  try {
    const url = adminDebugApi.blockProfile()
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.block = true
      pprofResults.value.blockText = text
      message.success(t('adminSettings.blockProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.block = false
  }
}

async function captureMutexProfile() {
  pprofLoading.mutex = true
  pprofResults.value.mutex = false
  pprofResults.value.mutexText = ''
  try {
    const url = adminDebugApi.mutexProfile()
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.mutex = true
      pprofResults.value.mutexText = text
      message.success(t('adminSettings.mutexProfileCompleted'))
    }
    else {
      message.error(t('adminSettings.captureFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.captureFailed') + e.message)
  }
  finally {
    pprofLoading.mutex = false
  }
}

async function loadRuntimeStacks() {
  loadingRuntimeStacks.value = true
  runtimeStackText.value = ''
  try {
    const url = adminDebugApi.goroutineProfile(stackFilterMinWaitMinutes.value)
    const token = authStorage.get('accessToken')
    const res = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })
    if (res.ok) {
      const text = await res.text()
      runtimeStackText.value = text
      const filterMsg = stackFilterMinWaitMinutes.value > 0 ? t('adminSettings.filtered', { minutes: stackFilterMinWaitMinutes.value }) : ''
      message.success(t('adminSettings.stacksLoaded') + filterMsg)
    }
    else {
      message.error(t('adminSettings.loadFailed'))
    }
  }
  catch (e: any) {
    message.error(t('adminSettings.loadFailed') + e.message)
  }
  finally {
    loadingRuntimeStacks.value = false
  }
}

function clearRuntimeStacks() {
  runtimeStackText.value = ''
  message.success(t('adminSettings.stacksCleared'))
}

onMounted(() => {
  loadSettings()
  loadServerMonitoringStatus()
})

onUnmounted(() => {
  if (debugRefreshInterval.value) {
    clearInterval(debugRefreshInterval.value)
  }
})
</script>

<style scoped></style>
