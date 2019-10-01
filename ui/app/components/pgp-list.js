import { computed } from '@ember/object';
import Component from '@ember/component';

export default Component.extend({
  onDataUpdate: () => {},
  listLength: 0,
  listData: null,

  init() {
    this._super(...arguments);
    let num = this.listLength;
    if (num) {
      num = parseInt(num, 10);
    }
    let list = this.newList(num);
    this.set('listData', list);
  },

  didReceiveAttrs() {
    this._super(...arguments);
    let list;
    if (!this.listLength) {
      this.set('listData', []);
      return;
    }
    // no update needed
    if (this.listData.length === this.listLength) {
      return;
    }
    // shorten the current list
    if (this.listLength < this.listData.length) {
      list = this.listData.slice(0, this.listLength);
    }
    // add to the current list by creating a new list and copying over existing list
    if (this.listLength > this.listData.length) {
      list = this.newList(this.listLength);
      if (this.listData.length) {
        list.splice(0, this.listData.length, ...this.listData);
      }
    }
    this.set('listData', list || this.listData);
    this.onDataUpdate((list || this.listData).compact().map(k => k.value));
  },

  newList(length) {
    return Array(length || 0)
      .fill(null)
      .map(() => ({ value: '' }));
  },

  listData: computed('listLength', function() {
    let num = this.get('listLength');
    if (num) {
      num = parseInt(num, 10);
    }
    return Array(num || 0)
      .fill(null)
      .map(() => ({ value: '' }));
  }),

  actions: {
    setKey(index, key) {
      let { listData } = this;
      listData.splice(index, 1, key);
      this.onDataUpdate(listData.compact().map(k => k.value));
    },
  },
});
