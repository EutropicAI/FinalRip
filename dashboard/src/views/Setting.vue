<script setup lang="ts">
import type { SelectOption } from 'naive-ui'
import type { Ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'
import { useSettingStore } from '@/store/setting'

const {
  apiURL,
  apiToken,
  templateRepo,
  githubToken,
  sliceMode,
  clipTimeout,
  queueName,
} = storeToRefs(useSettingStore())

const queueList: Ref<SelectOption[]> = ref([
  {
    label: 'default',
    value: 'default',
  },
  {
    label: 'priority',
    value: 'priority',
  },
])
</script>

<template>
  <NSpace vertical>
    <NCard hoverable title="Server API" size="small">
      <NSpace vertical>
        <NSpace>
          <div class="w-[50px]">
            URL:
          </div>
          <NInput v-model:value="apiURL" style="width: 40vh" placeholder="http://localhost:8848" />
        </NSpace>
        <NSpace>
          <div class="w-[50px]">
            Token:
          </div>
          <NInput
            v-model:value="apiToken"
            style="width: 40vh"
            type="password"
            show-password-on="click"
            placeholder="xxxxxx"
          />
        </NSpace>
      </NSpace>
    </NCard>
    <NCard hoverable title="Github API" size="small">
      <NSpace>
        <div class="w-[50px]">
          Token:
        </div>
        <NInput
          v-model:value="githubToken"
          style="width: 40vh"
          type="password"
          show-password-on="click"
          placeholder="ghp_ddddddddd1111111114514"
        />
      </NSpace>
    </NCard>
    <NCard hoverable title="Template Repo" size="small">
      <NSpace>
        <div class="w-[50px]">
          Name:
        </div>
        <NInput
          v-model:value="templateRepo"
          style="width: 40vh"
          placeholder="EutropicAI/vs-playground"
        />
      </NSpace>
    </NCard>
    <NCard hoverable title="Clip" size="small">
      <NSpace vertical>
        <NSpace>
          <div class="w-[150px]">
            Cut Video into clips:
          </div>
          <NSwitch v-model:value="sliceMode" size="large" />
        </NSpace>
        <NSpace>
          <div class="w-[150px]">
            Timeout:
          </div>
          <NInputNumber
            v-model:value="clipTimeout"
            :precision="0"
            placeholder="clip timeout, minute"
            :min="1"
            style="width: 20vh"
          />
        </NSpace>
      </NSpace>
    </NCard>
    <NCard hoverable title="Queue" size="small">
      <NSpace>
        <div class="w-[150px]">
          Encode Queue:
        </div>
        <NSelect v-model:value="queueName" :options="queueList" style="width: 20vh" />
      </NSpace>
    </NCard>
  </NSpace>
</template>

<style scoped></style>
