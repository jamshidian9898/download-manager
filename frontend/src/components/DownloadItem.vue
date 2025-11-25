<script setup>
import { computed } from 'vue'
import { 
  Play,
  Pause,
  Square,
  Trash2,
  Info,
  MoreHorizontal,
  CheckCircle,
  XCircle,
  Clock,
  Download as DownloadIcon,
  StopCircle
} from 'lucide-vue-next'

import { formatters } from '../utils/api.js'

const props = defineProps({
  download: {
    type: Object,
    required: true
  },
  selected: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'select', 
  'show-details', 
  'start', 
  'pause', 
  'resume', 
  'cancel', 
  'delete'
])

const statusIcon = computed(() => {
  const icons = {
    'pending': Clock,
    'downloading': DownloadIcon,
    'paused': Pause,
    'completed': CheckCircle,
    'failed': XCircle,
    'canceled': StopCircle
  }
  return icons[props.download.status] || Clock
})

const statusColor = computed(() => {
  const colors = {
    'pending': 'text-yellow-500',
    'downloading': 'text-blue-500',
    'paused': 'text-orange-500',
    'completed': 'text-green-500',
    'failed': 'text-red-500',
    'canceled': 'text-gray-500'
  }
  return colors[props.download.status] || 'text-gray-500'
})

const progressPercentage = computed(() => {
  if (!props.download.total_size || props.download.total_size === 0) return 0
  return Math.round((props.download.progress?.bytes_downloaded || 0) / props.download.total_size * 100)
})

const progressBarColor = computed(() => {
  switch (props.download.status) {
    case 'downloading':
      return 'bg-blue-500'
    case 'completed':
      return 'bg-green-500'
    case 'paused':
      return 'bg-orange-500'
    case 'failed':
      return 'bg-red-500'
    default:
      return 'bg-gray-500'
  }
})

const estimatedTimeLeft = computed(() => {
  if (props.download.status !== 'downloading' || !props.download.progress?.speed) {
    return '--'
  }
  
  const remaining = props.download.total_size - (props.download.progress?.bytes_downloaded || 0)
  const seconds = remaining / props.download.progress.speed
  return formatters.formatDuration(seconds)
})

const canStart = computed(() => {
  return ['pending', 'paused', 'failed'].includes(props.download.status)
})

const canPause = computed(() => {
  return props.download.status === 'downloading'
})

const canResume = computed(() => {
  return props.download.status === 'paused'
})

const formatDate = (dateString) => {
  if (!dateString) return '--'
  const date = new Date(dateString)
  const now = new Date()
  const diffHours = Math.floor((now - date) / (1000 * 60 * 60))
  
  if (diffHours < 1) {
    return 'Just now'
  } else if (diffHours < 24) {
    return `${diffHours}h ago`
  } else {
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays}d ago`
  }
}

const handleAction = (action) => {
  emit(action, props.download.id)
}
</script>

<template>
  <div 
    :class="[
      'group border-b border-dark-700 px-4 py-3 hover:bg-dark-800/50 transition-colors',
      selected ? 'bg-primary-600/20' : ''
    ]"
  >
    <div class="grid grid-cols-12 gap-4 items-center text-sm">
      <!-- Checkbox -->
      <div class="col-span-1">
        <input
          type="checkbox"
          :checked="selected"
          @change="emit('select')"
          class="rounded border-gray-600 bg-dark-700 text-primary-600 focus:ring-primary-500"
        />
      </div>
      
      <!-- Name with Icon -->
      <div class="col-span-4">
        <div class="flex items-center space-x-3">
          <component 
            :is="statusIcon" 
            :class="['w-4 h-4 flex-shrink-0', statusColor]"
          />
          <div class="min-w-0 flex-1">
            <div class="font-medium text-white truncate" :title="download.filename">
              {{ download.filename }}
            </div>
            <div class="text-xs text-gray-400 truncate" :title="download.url">
              {{ download.url }}
            </div>
          </div>
        </div>
      </div>
      
      <!-- Size -->
      <div class="col-span-1 text-gray-300">
        {{ formatters.formatBytes(download.total_size || 0) }}
      </div>
      
      <!-- Status -->
      <div class="col-span-1">
        <span :class="['capitalize font-medium', statusColor]">
          {{ download.status }}
        </span>
      </div>
      
      <!-- Downloaded with Progress Bar -->
      <div class="col-span-2">
        <div class="space-y-1">
          <div class="flex justify-between text-xs">
            <span class="text-gray-300">
              {{ formatters.formatBytes(download.progress?.bytes_downloaded || 0) }}
            </span>
            <span class="text-gray-400">
              {{ progressPercentage }}%
            </span>
          </div>
          <div class="w-full bg-dark-600 rounded-full h-2">
            <div 
              :class="['h-2 rounded-full transition-all duration-300', progressBarColor]"
              :style="{ width: `${progressPercentage}%` }"
            ></div>
          </div>
        </div>
      </div>
      
      <!-- Speed -->
      <div class="col-span-1 text-gray-300">
        <span v-if="download.status === 'downloading' && download.progress?.speed">
          {{ formatters.formatSpeed(download.progress.speed) }}
        </span>
        <span v-else class="text-gray-500">--</span>
      </div>
      
      <!-- Time Left -->
      <div class="col-span-1 text-gray-300">
        {{ estimatedTimeLeft }}
      </div>
      
      <!-- Date Added -->
      <div class="col-span-1 text-gray-400 text-xs">
        {{ formatDate(download.created_at) }}
      </div>
    </div>
    
    <!-- Action Buttons (shown on hover) -->
    <div class="flex items-center justify-end space-x-1 mt-2 opacity-0 group-hover:opacity-100 transition-opacity">
      <button
        v-if="canStart"
        @click="handleAction('start')"
        class="p-1 text-gray-400 hover:text-green-500 transition-colors"
        title="Start Download"
      >
        <Play class="w-4 h-4" />
      </button>
      
      <button
        v-if="canPause"
        @click="handleAction('pause')"
        class="p-1 text-gray-400 hover:text-orange-500 transition-colors"
        title="Pause Download"
      >
        <Pause class="w-4 h-4" />
      </button>
      
      <button
        v-if="canResume"
        @click="handleAction('resume')"
        class="p-1 text-gray-400 hover:text-blue-500 transition-colors"
        title="Resume Download"
      >
        <Play class="w-4 h-4" />
      </button>
      
      <button
        @click="handleAction('cancel')"
        class="p-1 text-gray-400 hover:text-red-500 transition-colors"
        title="Cancel Download"
      >
        <Square class="w-4 h-4" />
      </button>
      
      <button
        @click="emit('show-details')"
        class="p-1 text-gray-400 hover:text-blue-500 transition-colors"
        title="Show Details"
      >
        <Info class="w-4 h-4" />
      </button>
      
      <button
        @click="handleAction('delete')"
        class="p-1 text-gray-400 hover:text-red-500 transition-colors"
        title="Delete Download"
      >
        <Trash2 class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.group:hover .group-hover\:opacity-100 {
  opacity: 1;
}
</style>
