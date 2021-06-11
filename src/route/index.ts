function getCurrentHost(): string {
  return 'localhost';
}
function getCurrentPort(): string {
  return '8888';
}
function getCurrentProto(): string {
  return 'http';
}
function getFullHostSpec(): string {
  return `${getCurrentProto()}://${getCurrentHost()}:${getCurrentPort()}`;
}
const s5Asset = (name: string): string => {
  return `${getFullHostSpec()}/${name}`;
};
const appAsset = (name: string): string => {
  return `${getFullHostSpec()}/app/${name}`;
};

export {
  s5Asset as S5Asset,
  appAsset as AppAsset,
}
