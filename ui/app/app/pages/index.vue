<script setup lang="ts">
const localePath = useLocalePath()
await navigateTo(localePath('/dashboard'), { redirectCode: 302 })
</script>

<template>
  <div />
</template>
