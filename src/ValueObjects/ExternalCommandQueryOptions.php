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


class ExternalCommandQueryOptions extends QueryOptions {

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $servicedescription;

    /**
     * @var string
     */
    private $nodeName;

    /**
     * @var int
     */
    private $command;

    /**
     * @var int
     */
    private $end = 0;

    /**
     * ExternalCommandQueryOptions constructor.
     * @param array $params
     */
    public function __construct($params) {
        parent::__construct($params);
        $this->params = $params;

        if (isset($this->params['hostname'])) {
            $this->hostname = $this->params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->servicedescription = $params['servicedescription'];
        }

        if (isset($this->params['node_name'])) {
            $this->nodeName = $this->params['node_name'];
        }

        if (isset($this->params['command'])) {
            $this->command = $this->params['command'];
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
    public function getServicedescription() {
        return $this->servicedescription;
    }

    /**
     * @return int
     */
    public function getCommand() {
        return (int)$this->command;
    }

    /**
     * @return string
     */
    public function getNodeName() {
        return $this->nodeName;
    }

}
