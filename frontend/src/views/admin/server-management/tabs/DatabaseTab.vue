<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import DatabaseObjectTree from '../database/DatabaseObjectTree.vue'
import DdlPane from '../database/DdlPane.vue'
import SqlConsolePane from '../database/SqlConsolePane.vue'
import TableDataPane from '../database/TableDataPane.vue'
import TableStructurePane from '../database/TableStructurePane.vue'
import { useDatabaseConsole } from '../database/useDatabaseConsole'

const { t } = useI18n()
const consoleState = reactive(useDatabaseConsole())

function changePage(page: number) {
  consoleState.query.page = page
  void consoleState.loadRows()
}

function changePageSize(pageSize: number) {
  consoleState.query.page_size = pageSize
  consoleState.query.page = 1
  void consoleState.loadRows()
}

onMounted(() => {
  void consoleState.loadInfoAndTables()
})
</script>

<template>
  <NCard :bordered="false" size="small" class="database-console-card">
    <template #header>
      <NSpace align="center">
        <NText strong>
          {{ t('adminServer.dbTab') }}
        </NText>
        <NTag type="info" size="small">
          {{ t('adminServer.dbDriver') }}: {{ consoleState.driver || '-' }}
        </NTag>
        <NTag :type="consoleState.writeEnabled ? 'warning' : 'success'" size="small">
          {{ consoleState.writeEnabled ? t('adminServer.dbWriteMode') : t('adminServer.dbReadOnlyMode') }}
        </NTag>
      </NSpace>
    </template>
    <template #header-extra>
      <NSpace>
        <NButton
          v-if="consoleState.backupSupported"
          size="small"
          secondary
          type="primary"
          :loading="consoleState.backupLoading"
          @click="consoleState.downloadBackup"
        >
          {{ t('adminServer.dbBackup') }}
        </NButton>
        <NButton size="small" :loading="consoleState.loading" @click="consoleState.loadInfoAndTables">
          {{ t('common.refresh') }}
        </NButton>
      </NSpace>
    </template>

    <NAlert type="info" :bordered="false" class="mb-12px">
      {{ t('adminServer.dbConsoleHint') }}
    </NAlert>

    <NSplit class="database-console-split" :default-size="0.22" :min="0.16" :max="0.36">
      <template #1>
        <DatabaseObjectTree
          :loading="consoleState.loading"
          :model-value="consoleState.selectedTable"
          :tables="consoleState.tables"
          @refresh="consoleState.loadInfoAndTables"
          @update:model-value="consoleState.selectTable"
        />
      </template>
      <template #2>
        <div class="database-console-main">
          <NEmpty v-if="!consoleState.selectedTable" :description="t('adminServer.dbSelectTable')" />
          <template v-else>
            <NSpace align="center" class="mb-12px">
              <NText depth="3">
                {{ t('adminServer.dbCurrentObject') }}
              </NText>
              <NText strong>
                {{ consoleState.selectedTableLabel }}
              </NText>
            </NSpace>
            <NTabs type="line" animated>
              <NTabPane name="data" :tab="t('adminServer.dbData')">
                <TableDataPane
                  :columns="consoleState.rowColumns"
                  :create-row="consoleState.createRow"
                  :delete-row="consoleState.deleteRow"
                  :loading="consoleState.loading"
                  :meta-columns="consoleState.meta?.columns || []"
                  :page="consoleState.query.page"
                  :page-size="consoleState.query.page_size"
                  :rows="consoleState.rows"
                  :total="consoleState.query.total"
                  :update-row="consoleState.updateRow"
                  :write-enabled="consoleState.writeEnabled"
                  @page-change="changePage"
                  @page-size-change="changePageSize"
                />
              </NTabPane>
              <NTabPane name="structure" :tab="t('adminServer.dbStructure')">
                <TableStructurePane :loading="consoleState.metaLoading" :meta="consoleState.meta" />
              </NTabPane>
              <NTabPane name="sql" :tab="t('adminServer.dbSqlTitle')">
                <SqlConsolePane :selected-table="consoleState.selectedTable" :write-enabled="consoleState.writeEnabled" />
              </NTabPane>
              <NTabPane name="ddl" :tab="t('adminServer.dbDdl')">
                <DdlPane :ddl="consoleState.ddl" :loading="consoleState.ddlLoading" @refresh="consoleState.loadDdl" />
              </NTabPane>
            </NTabs>
          </template>
        </div>
      </template>
    </NSplit>
  </NCard>
</template>

<style scoped>
.database-console-split {
  min-height: 620px;
}

.database-console-main {
  min-width: 0;
  padding-left: 20px;
}

@media (max-width: 768px) {
  .database-console-split {
    min-height: 700px;
  }

  .database-console-main {
    padding-left: 12px;
  }
}
</style>
