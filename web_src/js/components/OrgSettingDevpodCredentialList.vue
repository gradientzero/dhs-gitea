<script>

import {createApp} from 'vue';
import {SvgIcon} from '../svg.js';

const sfc = {
  components: {SvgIcon},
  data(){
    return {
      credentials: window.config.settingsDevpodCredentialData.credentials,
      csrfToken: window.config.settingsDevpodCredentialData.csrfToken,
      link: window.config.settingsDevpodCredentialData.link,
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
    editLink(credential){
      return this.$data.link + '/edit?id=' + credential.ID
    },
    deleteCredential(event) {
      event.preventDefault();

      $('#credential-delete-modal')
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

export function initOrgSettingDevpodCredentialList(){
  const el = document.getElementById('setting-devpod-credential-app');
  if (!el) return;

  const view = createApp(sfc);
  view.mount(el);
}
</script>
<template>

    <div v-if="!credentials.length">
      No Credential has been setup.
    </div>

    <div v-else class="flex-list">
      <div class="flex-item" v-for="credential in credentials">

        <div class="flex-item-leading text green">
          <SvgIcon name="octicon-key" :size=32 />
        </div>

        <div class="flex-item-main">
          <div class="flex-item-title">{{credential.Name}}</div>
          <div class="flex-item-body">
            {{credential.Remote}} <br />
            {{credential.Key}} = {{obfuscate(credential.Value)}}
          </div>
        </div>

        <div class="flex-item-trailing">
          <a :href="editLink(credential)" class="ui tiny button">
            <SvgIcon name="octicon-pencil" />
          </a>
          <form class="ui form" :action="deleteLink" method="post" @submit="deleteCredential($event)">
            <input type="hidden" name="_csrf" :value="csrfToken">
            <input name="id" type="hidden" :value="credential.ID">
            <button type="submit" class="ui red tiny button">
              <SvgIcon name="octicon-trash" />
            </button>
          </form>
        </div>
      </div>
    </div>


  <!--      Modal to delete token-->
  <div class="ui modal" id="credential-delete-modal">
    <div class="header">Devpod Credential Delete</div>
    <div class="content">
      <p>Are you sure to delete this credential?</p>
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
