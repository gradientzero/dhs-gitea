<script>

import {createApp} from 'vue';
import {SvgIcon} from '../svg.js';

const sfc = {
  components: {SvgIcon},
  data(){
    return {
      tokens: window.config.settingsGiteaTokenData.tokens,
      csrfToken: window.config.settingsGiteaTokenData.csrfToken,
      link: window.config.settingsGiteaTokenData.link,
    };
  },
  computed:{
    deleteLink() {
      return this.$data.link + '/delete';
    }
  },
  methods: {
    obfuscate(str) {
      if (str.length <= 5) {
        return '*'.repeat(str.length || 0);
      }
      return str.slice(0,5) + '*'.repeat(str.length - 5 || 0);
    },
    deleteToken(event) {
      event.preventDefault();

      $('#token-delete-modal')
        .modal({
          onApprove: function() {
            event.target.submit();
          }
        })
        .modal('show');
    }
  }
};

export default sfc;

export function initOrgSettingGiteaTokenList(){
  const el = document.getElementById('setting-gitea-token-app');
  if (!el) return;

  const view = createApp(sfc);
  view.mount(el);
}
</script>
<template>

    <div v-if="!tokens.length">
      No token has been setup.
    </div>

    <div v-else class="flex-list">
      <div class="flex-item" v-for="token in tokens">

        <div class="flex-item-leading text green">
          <SvgIcon name="octicon-key" :size=32 />
        </div>

        <div class="flex-item-main">
          <div class="flex-item-title">{{token.Name}}</div>
          <div class="flex-item-body">
            {{ obfuscate(token.Token) }}
          </div>
        </div>

        <div class="flex-item-trailing">
          <form class="ui form" :action="deleteLink" method="post" @submit="deleteToken($event)">
            <input type="hidden" name="_csrf" :value="csrfToken">
            <input name="id" type="hidden" :value="token.ID">
            <button type="submit" class="ui red tiny button">
              <SvgIcon name="octicon-trash" />
            </button>
          </form>
        </div>
      </div>
    </div>

    <!--      Modal to delete token-->
    <div class="ui modal" id="token-delete-modal">
      <div class="header">Gitea Token Delete</div>
      <div class="content">
        <p>Are you sure to delete this token?</p>
      </div>
      <div class="actions">
        <div class="ui positive button">Delete</div>
        <div class="ui negative button">Cancel</div>
    </div>
  </div>
</template>
<style scoped>
.flex-list > .flex-item:only-child {
  padding-bottom: 0;
}
</style>
