<template>
  <div class="no-permission-page">
    <div class="page-navbar">
      <NavBar />
    </div>
    <div class="content">
      <a-result
        class="result"
        status="403"
        subtitle="您的账号暂未分配任何菜单权限，请联系管理员开通权限后重新登录"
      >
      </a-result>
      <div class="operation-row">
        <a-button key="logout" type="primary" @click="handleLogout">
          退出登录
        </a-button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { provide } from 'vue';
  import NavBar from '@/components/navbar/index.vue';
  import useUser from '@/hooks/user';

  const { logout } = useUser();

  // NavBar 依赖该注入用于移动端展开侧边菜单，此页面无侧边菜单，提供空实现即可
  provide('toggleDrawerMenu', () => {});

  const handleLogout = () => {
    logout();
  };
</script>

<style scoped lang="less">
  .no-permission-page {
    width: 100%;
    height: 100%;
  }

  .page-navbar {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 100;
    width: 100%;
    height: 60px;
  }

  .content {
    position: absolute;
    top: 50%;
    left: 50%;
    margin-left: -160px;
    margin-top: -121px;
    text-align: center;
  }

  .operation-row {
    margin-top: 16px;
  }
</style>
