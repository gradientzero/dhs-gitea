<script>

import {createApp} from 'vue';
import {SvgIcon} from '../svg.js';

const sfc = {
  components: {SvgIcon},
  data(){
    return {
      sshKeys: window.config.settingSshData.sshKeys,
      csrfToken: window.config.settingSshData.csrfToken,
      link: window.config.settingSshData.link,
      selectKey: {}, // empty by default
    };
  },
  computed:{
    deleteLink() {
      return this.$data.link + '/delete';
    }
  },
  methods: {

    showKey(key){
      this.$data.selectKey = key;
      $('#ssh-modal').modal('show');
    },

    deleteKey(event) {
      event.preventDefault();

      // Ref: https://stackoverflow.com/questions/25616832/how-to-use-custom-callback-in-semantic-ui-modal
      $('#ssh-delete-modal')
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

export function initOrgSettingSshKeyList(){
  const el = document.getElementById('setting-ssh-key-app');
  if (!el) return;

  const view = createApp(sfc);
  view.mount(el);
}
</script>
<template>

    <div v-if="!sshKeys.length">
      No key has been setup yet.
    </div>

    <div v-else class="flex-list">
      <div class="flex-item" v-for="key in sshKeys">

        <div class="flex-item-leading text green">
          <SvgIcon name="octicon-key" :size=32 />
        </div>

        <div class="flex-item-main">
          <div class="flex-item-title">{{key.Name}}</div>
          <div class="flex-item-body">
            {{ key.Fingerprint }}
          </div>
        </div>

        <div class="flex-item-trailing">
          <a class="ui primary tiny button" @click="showKey(key)">
            <SvgIcon name="octicon-search" />
          </a>
          <form class="ui form" :action="deleteLink" method="post" @submit="deleteKey($event)">
            <input type="hidden" name="_csrf" :value="csrfToken">
            <input name="id" type="hidden" :value="key.ID">
            <button type="submit" class="ui red tiny button">
              <SvgIcon name="octicon-trash" />
            </button>
          </form>
        </div>
      </div>
    </div>

  <!--      Modal to show pubic key-->
  <div id="ssh-modal" class="ui modal">
    <div class="header ui">{{selectKey.Name}}</div>
    <div class="content ui">
      <p>{{selectKey.PublicKey}}</p>
    </div>
    <div class="actions">
      <button class="ui black deny button">
        Close
      </button>
    </div>
  </div>

  <!--      Modal to delete key-->
  <div class="ui modal" id="ssh-delete-modal">
    <div class="header">Ssh Key Delete</div>
    <div class="content">
      <p>Are you sure to delete key?</p>
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
