import * as S5Err from '../error';
import { eventDescriptor } from './eventdescriptor';
import s5event from './s5event';

// xxx fixme(iansmith): using _string_ here in this type is uber crappy
// xxx fixme(iansmith): but typescript doesn't have extensible enums
export default class eventGenerator extends Map<string, eventDescriptor> {
  Generate(raw: Event | null, et: string): s5event {
    const desc = this.get(et);
    if (desc === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.CantCreate,
        `bad event type name:${et}`
      );
    }
    let x: number | undefined;
    let y: number | undefined;
    if (raw && 'clientX' in raw && 'clientY' in raw) {
      type hasCoords = {
        clientX: number;
        clientY: number;
      };
      x = (raw as hasCoords).clientX;
      y = (raw as hasCoords).clientY;
    }
    const result = desc.Generate(raw, x, y);
    if (raw) {
      raw.preventDefault();
    }
    return result;
  }
}
