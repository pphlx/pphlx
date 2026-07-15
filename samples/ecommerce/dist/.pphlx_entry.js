import StarRating from './components/StarRating.solid.jsx';
window.StarRating = StarRating;
import BuyButton from './components/BuyButton.svelte';
window.BuyButton = BuyButton;
import { render as solidRender } from 'solid-js/web';
window.SolidJS = { render: solidRender };
import { mount as svelteMount } from 'svelte';
window.Svelte = { mount: svelteMount };
