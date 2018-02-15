<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\ValueObjects;


class LogentryQueryOptions extends QueryOptions {

    /**
     * @var int
     */
    private $entry_time__gt = 0;

    /**
     * @var int
     */
    private $entry_time__lt = 0;

    /**
     * @var string
     */
    private $logentry_data__like = '';

    /**
     * LogentryQueryOptions constructor.
     * @param array $params
     */
    public function __construct($params) {
        parent::__construct($params);
        $this->params = $params;

        if (isset($this->params['entry_time__gt'])) {
            $this->entry_time__gt = (int)$this->params['entry_time__gt'];
        }

        if (isset($this->params['entry_time__lt'])) {
            $this->entry_time__lt = (int)$this->params['entry_time__lt'];
        }

        if (isset($this->params['logentry_data__like'])) {
            $this->logentry_data__like = $this->params['logentry_data__like'];
        }

    }

    /**
     * @return int
     */
    public function getEntryTimeGt() {
        return $this->entry_time__gt;
    }

    /**
     * @return int
     */
    public function getEntryTimeLt() {
        return $this->entry_time__lt;
    }

    /**
     * @return string
     */
    public function getLogentryDataLike() {
        return $this->logentry_data__like;
    }
}
