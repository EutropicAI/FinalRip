import type { SelectOption } from 'naive-ui'
import type { Ref } from 'vue'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type DarkModeType = 'system' | 'light' | 'dark'

export const useSettingStore = defineStore(
  'GlobalSetting_250902',
  () => {
    // dark mode
    const darkMode: Ref<DarkModeType> = ref('system')
    const systemDarkMode = ref(false)

    // menu collapsed
    const collapsed = ref(false)

    // api
    const apiURL = ref('')
    const apiToken = ref('114514')

    // templates
    const vsScriptTemplates: Ref<SelectOption[]> = ref([])
    const ffmpegParamTemplates: Ref<SelectOption[]> = ref([])
    const templateRepo = ref('EutropicAI/vs-playground')
    const githubToken = ref('')

    // encode
    const script = ref(
      'import os\n'
      + 'import vapoursynth as vs\n'
      + 'from vapoursynth import core\n'
      + '\n'
      + 'clip = core.bs.VideoSource(source=os.getenv(\'FINALRIP_SOURCE\'))\n'
      + 'clip.set_output()'
      + '\n',
    )
    const encodeParam = ref('ffmpeg -i - -vcodec libx265 -crf 16')

    // clip setting
    const sliceMode = ref(true)
    const clipTimeout = ref(20)

    // queue setting
    const queueName = ref('default')

    // list setting
    const checkedPendingBox = ref(false)
    const checkedRunningBox = ref(true)
    const checkedCompletedBox = ref(true)

    return {
      darkMode,
      systemDarkMode,
      collapsed,
      apiURL,
      apiToken,
      vsScriptTemplates,
      ffmpegParamTemplates,
      templateRepo,
      githubToken,
      script,
      encodeParam,
      sliceMode,
      clipTimeout,
      queueName,
      checkedPendingBox,
      checkedRunningBox,
      checkedCompletedBox,
    }
  },
  {
    persist: {
      storage: localStorage,
    },
  },
)
