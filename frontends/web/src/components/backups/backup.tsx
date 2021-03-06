/**
 * Copyright 2018 Shift Devices AG
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Component, h, RenderableProps } from 'preact';
import { translate, TranslateProps } from '../../decorators/translate';
import { Radio } from '../forms';

interface BackupsListItemProps {
    backup: Backup;
    selectedBackup: string | undefined;
    handleChange: (value: string) => void;
    onFocus: ({ target }: { target: HTMLElement; }) => void;
}

export interface Backup {
    id: string;
    date: string;
    name: string;
}

type Props = BackupsListItemProps & TranslateProps;

class BackupsListItem extends Component<Props> {
    public render(
        { backup, selectedBackup, handleChange, onFocus }: RenderableProps<Props>,
    ) {
        let date = '';
        if (backup.date && backup.date !== '') {
            const options = {
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
            };
            date = new Date(backup.date).toLocaleString(this.context.i18n.language, options);
        } else {
            date = 'unknown';
        }
        return (
            <Radio
                checked={selectedBackup === backup.id}
                onChange={event => handleChange(event.target.value)}
                id={backup.id}
                label={backup.name && backup.name !== '' ? backup.name : backup.id}
                value={backup.id}
                onFocus={onFocus}
                sizeMedium>
                <span className="text-small text-gray">{date}</span>
            </Radio>
        );
    }
}

const TranslatedBackupsListItem = translate<BackupsListItemProps>()(BackupsListItem);
export { TranslatedBackupsListItem as BackupsListItem };
