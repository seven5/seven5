import attribute from './base';
import * as S5Conf from '../conf';
import basicAttribute from './basicattr';

let { IdCounter } = S5Conf.Vars;

export default class equalAttribute extends basicAttribute<number> {
  enforce = (v: number): number => {
    return v;
  };

  constructor(source: attribute<number>, name?: string | undefined) {
    super(name || `[equal ${IdCounter}]`);
    IdCounter += 1;
    this.SetInputsAndFunc([source], this.enforce);
    this.dirty = true; // because you have to compute it to have a cached copy
  }

  // eslint-disable-next-line class-methods-use-this
  AttributeName(): string {
    return `Equal`;
  }
}
