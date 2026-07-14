export default class PreactCart extends window.preact.Component {
  constructor(props) {
    super(props);
    this.state = { items: 2 };
  }
  render() {
    const { h } = window.preact;
    const price = this.props.price || 19.99;
    return h('div', { class: 'p-6 border border-[#673ab7] rounded-lg bg-gray-900 shadow-md' }, [
      h('h3', { class: 'text-xl font-bold text-[#673ab7] mb-3' }, '7. Preact: Cart Manager'),
      h('p', { class: 'text-xs text-gray-300 mb-2' }, `Item price: $${price}`),
      h('div', { class: 'flex items-center gap-3.5 mb-3.5' }, [
        h('button', {
          onClick: () => this.setState({ items: Math.max(0, this.state.items - 1) }),
          class: 'bg-gray-800 text-white border border-[#673ab7]/30 w-9 h-9 rounded text-base flex items-center justify-center cursor-pointer transition-all hover:bg-gray-700'
        }, '-'),
        h('span', { class: 'text-lg font-bold text-white w-6 text-center' }, this.state.items),
        h('button', {
          onClick: () => this.setState({ items: this.state.items + 1 }),
          class: 'bg-gray-800 text-white border border-[#673ab7]/30 w-9 h-9 rounded text-base flex items-center justify-center cursor-pointer transition-all hover:bg-gray-700'
        }, '+')
      ]),
      h('p', { class: 'text-xs font-bold text-gray-300 border-t border-gray-800 pt-3 mt-3' }, `Total Cost: $${(this.state.items * price).toFixed(2)}`)
    ]);
  }
}
