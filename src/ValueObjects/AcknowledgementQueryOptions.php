<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
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


class AcknowledgementQueryOptions extends QueryOptions {

    /**
     * @var bool
     */
    private $isHostRequest = true;

    /**
     * AcknowledgementQueryOptions constructor.
     * @param $params
     * @throws \Exception
     */
    public function __construct($params) {
        parent::__construct($params);

        $this->isHostRequest = true;
        if (isset($this->params['object_type'])) {
            if ($this->params['object_type'] == 'service') {
                $this->isHostRequest = false;
            }
        }

    }

    /**
     * @return bool
     */
    public function isHostRequest() {
        return $this->isHostRequest;
    }

}

