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

namespace Statusengine\Controller;


class Controller {

    /**
     * Make an php array javascript json compatible
     * that we don't loss order, or so
     * @param array $array
     * @return array
     */
    public function keepOrder($array = []) {
        $return = [];
        foreach ($array as $key => $value) {
            $return[] = [$key => $value];
        }
        return $return;
    }

    /**
     * @param int|array $states
     * @return int|array
     */
    public function getHostStateDescription($states = 0) {
        $descriptions = [
            0 => 'up',
            1 => 'down',
            2 => 'unreachable'
        ];

        if (is_array($states)) {
            $return = [];
            foreach ($states as $state => $count) {
                $return[$descriptions[$state]] = $count;
            }
            return $return;
        }

        if ($states > 2) {
            $states = 2;
        }
        return $descriptions[$states];

    }

    /**
     * @param int|array $states
     * @return int|array
     */
    public function getServiceStateDescription($states = 0) {
        $descriptions = [
            0 => 'ok',
            1 => 'warning',
            2 => 'critical',
            3 => 'unknown'
        ];

        if (is_array($states)) {
            $return = [];
            foreach ($states as $state => $count) {
                $return[$descriptions[$state]] = $count;
            }
            return $return;
        }

        if ($states > 2) {
            $states = 2;
        }
        return $descriptions[$states];

    }

}