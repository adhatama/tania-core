import NProgress from 'nprogress'

import * as types from '@/stores/mutation-types'
import API from '@/stores/api/farm'
import stub from '@/stores/stubs/user'

const state = {
  current: stub
}

const getters = {
  getCurrentUser: state => state.current,
  IsUserAuthenticated: state => state.current.uid != '',
  IsNewUser: (state, getters) => getters.haveFarms === false,
  IsUserAllowSeeNavigator: (state, getters) => {
    return getters.IsUserAuthenticated && getters.IsNewUser === false
  }
}

const actions = {
  userLogin ({ commit, state }, payload) {
    NProgress.start()
    return new Promise(( resolve, reject ) => {
      API
        .ApiLogin(payload).then(function(data) {
          commit(types.USER_LOGIN, {
            uid: 1,
            email: payload.email,
            intro: payload.email === 'user' ? false: true
          })
          resolve(data)
        }).catch(function() {
          reject('Incorrect Email and/or password')
        })
    })
  },
  userChangePassword ({ commit, state }, payload) {
    NProgress.start()
    return new Promise(( resolve, reject ) => {
      API
        .ApiChangePassword(payload, ({ data }) => {
          resolve(data)
        }, error => reject(error.response))
    })
  },
  userCompletedIntro({ commit, state }) {
    commit(types.USER_COMPLETED_INTRO)
  },
  userSignOut({commit, state}, payload) {
    return new Promise((resolve, reject) => {
      commit(types.USER_LOGOUT)
      resolve()
    })
  }
}

const mutations = {
  [types.USER_LOGIN] (state, { uid, email, intro }) {
    state.current = { uid, email, intro }
  },
  [types.USER_COMPLETED_INTRO] (state) {
    state.current.intro = false
  },
  [types.USER_LOGOUT] (state, payload) {
    state.current.uid = ''
  }
}

export default {
  state, getters, actions, mutations
}
