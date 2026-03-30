import { browser } from '$app/environment';
import { init, register, getLocaleFromNavigator, locale, isLoading } from 'svelte-i18n';
import en from './en.json';
import ru from './ru.json';

const defaultLocale = 'en';

register('en', () => Promise.resolve(en));
register('ru', () => Promise.resolve(ru));

init({
  fallbackLocale: defaultLocale,
  initialLocale: browser 
    ? (localStorage.getItem('langduel_locale') || getLocaleFromNavigator() || defaultLocale)
    : defaultLocale,
});

export { locale, isLoading };

export function setLocale(newLocale) {
  if (!browser) return;
  locale.set(newLocale);
  localStorage.setItem('langduel_locale', newLocale);
}
