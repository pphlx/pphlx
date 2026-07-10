export default class PreactCart extends window.preact.Component {
  constructor(props) {
    super(props);
    this.state = { items: 2 };
  }
  render() {
    const { h } = window.preact;
    const price = this.props.price || 19.99;
    return h('div', { style: 'padding: 20px; border: 1px solid #673ab7; border-radius: 8px; background: #1a1a24; color: #fff; margin: 15px 0;' }, [
      h('h3', { style: 'color: #673ab7; margin-top: 0;' }, 'Preact: Cart Manager'),
      h('p', null, `Item price: $${price}`),
      h('div', { style: 'display: flex; align-items: center; gap: 15px; margin-bottom: 10px;' }, [
        h('button', {
          onClick: () => this.setState({ items: Math.max(0, this.state.items - 1) }),
          style: 'background: #2d2d3d; color: #fff; border: 1px solid #673ab7; width: 36px; height: 36px; border-radius: 4px; cursor: pointer; font-size: 1.2em;'
        }, '-'),
        h('span', { style: 'font-size: 1.5em; font-weight: bold;' }, this.state.items),
        h('button', {
          onClick: () => this.setState({ items: this.state.items + 1 }),
          style: 'background: #2d2d3d; color: #fff; border: 1px solid #673ab7; width: 36px; height: 36px; border-radius: 4px; cursor: pointer; font-size: 1.2em;'
        }, '+')
      ]),
      h('p', { style: 'font-size: 1.1em; font-weight: bold; border-top: 1px solid #444; padding-top: 10px; margin-top: 10px;' }, `Total Cost: $${(this.state.items * price).toFixed(2)}`)
    ]);
  }
}
