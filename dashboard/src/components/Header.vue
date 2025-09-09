<script setup lang="ts">
import {
  ContrastSharp,
  LogoGithub,
  MoonOutline,
  SettingsOutline,
  SunnyOutline,
} from '@vicons/ionicons5'
import { storeToRefs } from 'pinia'
import router from '@/router'
import { useSettingStore } from '@/store/setting'

const { darkMode } = storeToRefs(useSettingStore())

class openWebsite {
  static async github(): Promise<void> {
    const githubLink = 'https://github.com/EutropicAI/FinalRip'

    window.open(githubLink, '_blank')
  }
}

function handleDarkMode(): void {
  if (darkMode.value === 'system') {
    darkMode.value = 'light'
  }
  else if (darkMode.value === 'light') {
    darkMode.value = 'dark'
  }
  else {
    darkMode.value = 'system'
  }
}
</script>

<template>
  <div class="mt-[-4px]">
    <NFlex justify="end">
      <NButton text size="large" @click="openWebsite.github()">
        <NIcon size="26">
          <LogoGithub />
        </NIcon>
      </NButton>

      <NButton text size="large" @click="handleDarkMode">
        <NIcon size="26">
          <div v-if="darkMode === 'light'">
            <SunnyOutline />
          </div>
          <div v-else-if="darkMode === 'dark'">
            <MoonOutline />
          </div>
          <div v-else>
            <ContrastSharp />
          </div>
        </NIcon>
      </NButton>

      <NButton text size="large" @click="router.push('/setting')">
        <NIcon size="26">
          <SettingsOutline />
        </NIcon>
      </NButton>
    </NFlex>
  </div>
</template>
