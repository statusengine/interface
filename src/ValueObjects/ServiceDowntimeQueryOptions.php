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


class ServiceDowntimeQueryOptions extends QueryOptions {

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $service_description;

    /**
     * ServiceCheckQueryOptions constructor.
     * @param $params
     * @throws \Exception
     */
    public function __construct($params) {
        parent::__construct($params);


        if (isset($this->params['hostname'])) {
            $this->hostname = $params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->service_description = $params['servicedescription'];
        }

    }


    /**
     * @return string
     */
    public function getHostname() {
        return $this->hostname;
    }

    /**
     * @return string
     */
    public function getServiceDescription() {
        return $this->service_description;
    }

}
