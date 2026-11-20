<template>
  <MemberListV2 v-if="layout === 'v2'" />
  <MemberListClassic v-else />
</template>

<script lang="ts" setup>
  import { useRoute } from 'vue-router';
  import MemberListClassic from './components/member-list-classic.vue';
  import MemberListV2 from './components/v2/member-list-v2.vue';
  import { useMemberLayout } from './layout';
  import { provideMemberStore } from './store';

  const route = useRoute();
  const layout = useMemberLayout(route.query.layout);
  provideMemberStore();
</script>
